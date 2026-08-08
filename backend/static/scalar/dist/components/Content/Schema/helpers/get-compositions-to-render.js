import { getRefName } from "./get-ref-name.js";
import { compositions } from "./schema-composition.js";
import { shouldRenderArrayItemComposition } from "./should-render-array-item-composition.js";
import { resolve } from "@scalar/workspace-store/resolve";
import { isDefined } from "@scalar/helpers/array/is-defined";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
//#region src/components/Content/Schema/helpers/get-compositions-to-render.ts
var normalizeDiscriminatorMappingRef = (value) => value.startsWith("#/") || value.includes("/") ? value : `#/components/schemas/${value}`;
/**
* Builds a synthetic `oneOf` composition from a `discriminator.mapping` when the
* schema declares a mapping but no explicit `oneOf`/`anyOf`. This is the shape
* NSwag emits for polymorphic types, where the base type is a plain object with
* a discriminator mapping pointing at the concrete variants.
*
* Returns `null` when there is nothing to infer (an explicit composition is
* already present, no document to resolve refs against, or no resolvable refs).
*/
var inferDiscriminatorMappingComposition = (value, document) => {
	if (value.oneOf || value.anyOf || !document?.components?.schemas) return null;
	const refs = Object.values(value.discriminator?.mapping ?? {}).filter((mappingValue) => typeof mappingValue === "string").map((mappingValue) => {
		const ref = normalizeDiscriminatorMappingRef(mappingValue);
		const refName = getRefName(ref);
		const refValue = refName ? resolve.schema(document.components?.schemas?.[refName]) : void 0;
		if (!refValue) return;
		return {
			$ref: ref,
			"$ref-value": refValue
		};
	}).filter(isDefined);
	if (refs.length === 0) return null;
	return {
		...resolve.schema(value),
		oneOf: refs
	};
};
/**
* Computes which compositions should be rendered and with which values
*
* @param value - The schema object to check for compositions
* @returns Array of compositions to render with their values
*/
var getCompositionsToRender = (value, document) => {
	if (!value) return [];
	const inferredDiscriminatorComposition = inferDiscriminatorMappingComposition(value, document);
	return compositions.map((composition) => {
		if (composition === "oneOf" && inferredDiscriminatorComposition) return {
			composition,
			value: inferredDiscriminatorComposition
		};
		if (shouldRenderArrayItemComposition(value, composition) && isArraySchema(value) && value.items) return {
			composition,
			value: resolve.schema(value.items)
		};
		if (value[composition]) {
			if (!(isArraySchema(value) && value.items && typeof value.items === "object" && composition in value.items)) return {
				composition,
				value: resolve.schema(value)
			};
		}
		return null;
	}).filter(isDefined);
};
//#endregion
export { getCompositionsToRender, inferDiscriminatorMappingComposition };

//# sourceMappingURL=get-compositions-to-render.js.map