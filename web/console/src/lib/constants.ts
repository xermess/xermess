/**
 * The panel's own settings, as opposed to the project's name: what the
 * project's name is, and the names built from it, are in brand.ts.
 */

/** The shortest password an administrator may have: the server's
    model.MinAdminPasswordLength, said here too so a form can say so before
    anything is sent. */
export const MIN_ADMIN_PASSWORD = 10;

/** What a load that reads the signed-in administrator depends on, so a
    change to their own account can refresh that alone rather than every
    load on the page. */
export const ADMIN_DEPENDENCY = 'app:admin';
