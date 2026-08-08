//#region src/components/Content/Schema/helpers/should-display-description.ts
/**
* Determine if description should be displayed
*
* @param schema - The schema object to check
* @param propDescription - Optional description from props
* @returns Description string to display, or null if should not be displayed
*/
var shouldDisplayDescription = (schema, propDescription) => {
	if (!schema) return null;
	if (schema.allOf) return null;
	if (propDescription && schema.description) return propDescription === schema.description ? propDescription : `${propDescription}\n\n${schema.description}`;
	return propDescription || schema.description || null;
};
//#endregion
export { shouldDisplayDescription };

//# sourceMappingURL=should-display-description.js.map