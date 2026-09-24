/**
 * Ark renders a switch's root as a <label>, and a label passes a click
 * anywhere inside it on to its input — so a click on the words, the
 * description, or the empty row between them turned the switch. Only the
 * switch itself should: a setting changed by a stray click is one nobody
 * meant to change.
 *
 * Given to the root's onclick, this cancels the label's part in any click
 * that did not land on the switch. Two clicks still go through: the one on
 * the switch, and the one the browser then sends the hidden input — which
 * is also how Space turns a focused switch from the keyboard.
 */
export function onlyTheSwitch(event: MouseEvent) {
	const target = event.target as Element;

	if (target instanceof HTMLInputElement || target.closest("[data-part='control']")) return;

	event.preventDefault();
}
