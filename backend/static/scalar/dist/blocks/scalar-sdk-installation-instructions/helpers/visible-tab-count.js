//#region src/blocks/scalar-sdk-installation-instructions/helpers/visible-tab-count.ts
/**
* Work out how many tabs fit inline given the width of each tab and the width
* available to render them in.
*
* When everything fits, all tabs are shown. Otherwise we reserve room for the
* "More" dropdown trigger and return how many tabs fit alongside it (always at
* least one, so the row is never empty).
*/
var getVisibleTabCount = (tabWidths, availableWidth, moreWidth) => {
	const total = tabWidths.length;
	if (availableWidth <= 0) return total;
	const countThatFits = (reserved) => {
		let used = 0;
		let count = 0;
		for (const width of tabWidths) {
			if (used + width > availableWidth - reserved) break;
			used += width;
			count++;
		}
		return count;
	};
	if (countThatFits(0) >= total) return total;
	return Math.max(1, countThatFits(moreWidth));
};
//#endregion
export { getVisibleTabCount };

//# sourceMappingURL=visible-tab-count.js.map