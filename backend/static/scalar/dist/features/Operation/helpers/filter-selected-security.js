import { getSecuritySchemes, getSelectedSecurity } from "@scalar/workspace-store/request-example";
//#region src/features/Operation/helpers/filter-selected-security.ts
/** Builds a quick cache key from the sorted object keys */
var getKey = (requirement) => Object.keys(requirement).sort().join(",");
/**
* Find the intersection between which security is selected on the document and what this operation requires
*
* If there is no overlap, we return the first requirement
*/
var filterSelectedSecurity = (document, operation, selectedSecurityDocument, selectedSecurityOperation, securitySchemes = {}) => {
	const securityRequirements = operation?.security ?? document.security ?? [];
	/** The selected security keys for the document */
	const selectedSecurity = getSelectedSecurity(selectedSecurityDocument, selectedSecurityOperation, securityRequirements);
	/** Build a set for O(1) lookup */
	const requirementSet = new Set(securityRequirements.map((r) => getKey(r)));
	const selectedRequirement = selectedSecurity.selectedSchemes[selectedSecurity.selectedIndex];
	if (selectedRequirement && requirementSet.has(getKey(selectedRequirement))) return getSecuritySchemes(securitySchemes, selectedRequirement);
	for (const selected of selectedSecurity.selectedSchemes) if (requirementSet.has(getKey(selected))) return getSecuritySchemes(securitySchemes, selected);
	/**
	* If we are selected security on the document,
	* we should show the first requirement of the operation to show auth is required
	*/
	if (operation?.security?.length) return getSecuritySchemes(securitySchemes, securityRequirements[0] ?? {});
	return [];
};
//#endregion
export { filterSelectedSecurity };

//# sourceMappingURL=filter-selected-security.js.map