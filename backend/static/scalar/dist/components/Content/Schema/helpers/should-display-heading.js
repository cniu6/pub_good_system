//#region src/components/Content/Schema/helpers/should-display-heading.ts
/**
* Determine if property heading should be displayed
*
* @param schema - The schema object to check
* @param name - Optional property name
* @param required - Whether the property is required
* @returns true if heading should be displayed, false otherwise
*/
var shouldDisplayHeading = (schema, name, required = false) => {
	if (name || required) return true;
	if (!schema) return false;
	return schema.deprecated === true || schema.const !== void 0 || schema.enum?.length === 1 || "type" in schema && schema.type !== void 0 || "nullable" in schema && schema.nullable === true || schema.writeOnly === true || schema.readOnly === true;
};
//#endregion
export { shouldDisplayHeading };

//# sourceMappingURL=should-display-heading.js.map