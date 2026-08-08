import { resolve } from "@scalar/workspace-store/resolve";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/components/Content/Schema/helpers/get-enum-values.ts
/**
* Extract enum values from schema or array items
*
* @param value - The schema object to extract enum values from
* @returns Array of enum values, or empty array if no enum found
*/
var getEnumValues = (value) => {
	if (!value) return [];
	if (value.enum) return value.enum;
	if (isArraySchema(value) && typeof value.items === "object") {
		const resolvedItems = resolve.schema(value.items);
		if (resolvedItems && "enum" in resolvedItems && resolvedItems.enum) return resolvedItems.enum;
	}
	return [];
};
//#endregion
export { getEnumValues };

//# sourceMappingURL=get-enum-values.js.map