import { hasComplexArrayItems } from "./has-complex-array-items.js";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/components/Content/Schema/helpers/should-render-array-item-composition.ts
/**
* Check if array item composition should be rendered
*
* @param schema - The schema object to check
* @param composition - The composition keyword to check for
* @returns true if array item composition should be rendered, false otherwise
*/
var shouldRenderArrayItemComposition = (schema, composition) => {
	if (!schema || !isArraySchema(schema)) return false;
	const items = schema.items;
	if (!items || typeof items !== "object" || !(composition in items)) return false;
	return !hasComplexArrayItems(schema);
};
//#endregion
export { shouldRenderArrayItemComposition };

//# sourceMappingURL=should-render-array-item-composition.js.map