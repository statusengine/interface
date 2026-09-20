/**
 * The two pieces of arithmetic behind the dashboard's percentages.
 *
 * Separate from the component because a percentage is where a dashboard
 * lies most easily - a divide by zero rendered as 0%, a change against
 * nothing rendered as +100% - and those are worth testing without a
 * component around them.
 */

/**
 * A share of a whole, in whole points, or null when there is nothing to
 * divide by.
 *
 * Whole points on purpose: 99.8% invites a question the data cannot
 * answer, because this is a snapshot of the current state and not an
 * availability measurement over time.
 */
export function share(part: number | undefined, whole: number | undefined): number | null {
  if (!whole) {
    return null;
  }
  return Math.round(((part ?? 0) / whole) * 100);
}

/**
 * The change from one period to the next, in whole points, or null when
 * the earlier period had nothing to change from.
 *
 * "Infinitely more than zero" is not a number a dashboard should print,
 * and "+100%" for the first alert ever would be worse: it looks like a
 * doubling.
 */
export function percentChange(now: number, before: number): number | null {
  if (before === 0) {
    return null;
  }
  return Math.round(((now - before) / before) * 100);
}
