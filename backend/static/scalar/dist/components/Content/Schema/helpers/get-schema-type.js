import { resolve } from "@scalar/workspace-store/resolve";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/components/Content/Schema/helpers/get-schema-type.ts
/**
* Formats an array type string with proper wrapping for union types.
*/
var formatArrayType = (itemType) => {
	if (!itemType) return "array";
	return `array ${itemType.includes(" | ") ? `(${itemType})` : itemType}[]`;
};
/**
* Handles array type processing for both single array types and union types containing array.
*/
var processArrayType = (value, isUnionType = false) => {
	if (!value.items) return "array";
	const baseType = formatArrayType(getSchemaType(resolve.schema(value.items)));
	if (isUnionType) return baseType;
	return value.nullable ? `${baseType} | null` : baseType;
};
/**
* Computes the structural type for a schema.
* This helper always returns type information, never schema titles or ref names.
*
* Priority order:
* 1. const values
* 2. Array types (with special handling for items)
* 3. type with contentEncoding
* 4. raw type
*/
var getSchemaType = (valueOrRef) => {
	if (!valueOrRef) return "";
	const value = resolve.schema(valueOrRef);
	if (value.const !== void 0) return "const";
	if ("type" in value && Array.isArray(value.type)) {
		if (value.type.includes("array") && value.items) {
			const arrayType = processArrayType(value, true);
			const otherTypes = value.type.filter((t) => t !== "array");
			return otherTypes.length > 0 ? `${arrayType} | ${otherTypes.join(" | ")}` : arrayType;
		}
		return value.type.join(" | ");
	}
	if (isArraySchema(value)) return processArrayType(value, false);
	if ("type" in value && value.type && value.contentEncoding) return `${value.type} • ${value.contentEncoding}`;
	return "type" in value ? value.type : "";
};
//#endregion
export { getSchemaType };

//# sourceMappingURL=get-schema-type.js.map