import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/components/Content/Schema/helpers/has-complex-array-items.ts
/** Composition keywords that indicate complex schema structure */
var COMPOSITION_KEYWORDS = [
	"allOf",
	"oneOf",
	"anyOf"
];
/**
* Checks if a schema has object type (either explicit type: 'object' or has properties)
*/
var isObjectType = (schema) => {
	if ("type" in schema && schema.type) {
		if (Array.isArray(schema.type)) return schema.type.includes("object");
		return schema.type === "object";
	}
	return "properties" in schema;
};
/**
* Checks if a schema has complex features (refs, compositions, discriminators)
*/
var hasComplexFeatures = (schema) => "$ref" in schema || "discriminator" in schema || COMPOSITION_KEYWORDS.some((keyword) => keyword in schema);
/**
* Checks if nested array items are complex
*/
var hasComplexNestedArrayItems = (items) => {
	if (!isArraySchema(items) || typeof items.items !== "object") return false;
	if ("$ref" in items.items) return true;
	const nestedItems = getResolvedRef(items.items);
	if (!nestedItems) return false;
	return isObjectType(nestedItems) || hasComplexFeatures(nestedItems) || isArraySchema(nestedItems);
};
/**
* Checks if array items have complex structure
* like: objects, references, discriminators, compositions, or nested arrays with complex items
*
* @param value - The schema object to check
* @returns true if the array has complex items, false otherwise
*/
var hasComplexArrayItems = (value) => {
	if (!value || !isArraySchema(value) || typeof value.items !== "object") return false;
	if ("$ref" in value.items) return true;
	const items = getResolvedRef(value.items);
	if (!items) return false;
	if (hasComplexFeatures(items)) return true;
	if (isObjectType(items)) return true;
	return hasComplexNestedArrayItems(items);
};
//#endregion
export { hasComplexArrayItems };

//# sourceMappingURL=has-complex-array-items.js.map