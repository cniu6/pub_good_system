import { getRefName } from "./get-ref-name.js";
import { resolve } from "@scalar/workspace-store/resolve";
//#region src/components/Content/Schema/helpers/schema-name.ts
/**
* Extract schema name from various schema formats
*
* Handles $ref, title, name, type, and schema dictionary lookup
*/
var getModelNameFromSchema = (schemaOrRef) => {
	if (!schemaOrRef) return null;
	const schema = resolve.schema(schemaOrRef);
	const schemaKey = "$ref" in schemaOrRef ? getRefName(schemaOrRef.$ref) ?? null : null;
	if (schema.title) return {
		schemaKey,
		label: schema.title
	};
	if (schema.name) return {
		schemaKey,
		label: schema.name
	};
	if ("$ref" in schemaOrRef) {
		if (schemaKey) return {
			schemaKey,
			label: schemaKey
		};
	}
	return null;
};
//#endregion
export { getModelNameFromSchema };

//# sourceMappingURL=schema-name.js.map