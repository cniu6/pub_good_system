//#region src/components/Content/Schema/helpers/is-type-object.ts
var isTypeObject = (schema) => {
	if (schema === null || typeof schema !== "object" || Array.isArray(schema)) return false;
	if ("oneOf" in schema || "anyOf" in schema || "allOf" in schema || "not" in schema) return false;
	const hasType = "type" in schema;
	if (hasType && Array.isArray(schema.type)) return schema.type.includes("object");
	const hasTypeObject = hasType && schema.type === "object";
	if (hasTypeObject) return true;
	if (hasType && !hasTypeObject) return false;
	const hasProperties = "properties" in schema;
	const hasAdditionalProperties = "additionalProperties" in schema;
	const hasPatternProperties = "patternProperties" in schema;
	return hasProperties || hasAdditionalProperties || hasPatternProperties;
};
//#endregion
export { isTypeObject };

//# sourceMappingURL=is-type-object.js.map