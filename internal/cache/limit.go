package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// takeToken is a token bucket kept in Redis, taken from in one atomic step so
// two server processes cannot both spend the last token.
//
// The bucket is a hash of how many tokens it held and when it was last
// touched, refilled for the time in between at `rate` tokens a second up to
// `burst`. The clock is Redis's own, so servers whose clocks disagree still
// agree on the bucket. It answers whether a token was taken, and when not,
// how many milliseconds until one will be there.
var takeToken = redis.NewScript(`
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local time = redis.call('TIME')
local now = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)

local bucket = redis.call('HMGET', KEYS[1], 'tokens', 'at')
local tokens = tonumber(bucket[1]) or burst
local at = tonumber(bucket[2]) or now

tokens = math.min(burst, tokens + math.max(0, now - at) / 1000 * rate)

local allowed, wait = 0, 0
if tokens >= 1 then
	tokens = tokens - 1
	allowed = 1
else
	wait = math.ceil((1 - tokens) / rate * 1000)
end

redis.call('HSET', KEYS[1], 'tokens', tostring(tokens), 'at', now)
redis.call('PEXPIRE', KEYS[1], math.ceil(burst / rate * 1000))

return {allowed, wait}
`)

// Take spends a token from the bucket named `bucket`, which allows
// `perMinute` a minute with the whole minute's worth available at once, and
// says whether there was one and, when not, how long until there is.
//
// An error means Redis did not answer; the caller decides what a limit is
// worth without it.
func (c *Cache) Take(ctx context.Context, bucket string, perMinute int) (bool, time.Duration, error) {
	if !c.available() {
		return false, 0, errUnavailable
	}

	rate := float64(perMinute) / 60

	result, err := takeToken.Run(ctx, c.client, []string{c.key("ratelimit", bucket)}, rate, perMinute).Int64Slice()
	if err != nil {
		c.failed(err)
		return false, 0, err
	}

	c.recovered()

	return result[0] == 1, time.Duration(result[1]) * time.Millisecond, nil
}
