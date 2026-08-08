/**
 * Work out how many tabs fit inline given the width of each tab and the width
 * available to render them in.
 *
 * When everything fits, all tabs are shown. Otherwise we reserve room for the
 * "More" dropdown trigger and return how many tabs fit alongside it (always at
 * least one, so the row is never empty).
 */
export declare const getVisibleTabCount: (
/** The natural width of each tab, in order */
tabWidths: number[], 
/** The width available to render the tabs in */
availableWidth: number, 
/** The width of the "More" dropdown trigger */
moreWidth: number) => number;
//# sourceMappingURL=visible-tab-count.d.ts.map