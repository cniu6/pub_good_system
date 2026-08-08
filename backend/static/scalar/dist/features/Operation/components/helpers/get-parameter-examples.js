import { isObjectLike } from "@scalar/helpers/object/is-object";
//#region src/features/Operation/components/helpers/get-parameter-examples.ts
var filterUndefined = (example) => example !== void 0;
/**
* Build a normalized examples array from parameter/content/schema examples.
* Undefined values are removed so the UI does not render "undefined" entries.
*/
var getParameterExamples = ({ parameter, schemaExamples, contentExamples }) => {
	const paramExamples = "examples" in parameter && isObjectLike(parameter.examples) ? parameter.examples : {};
	const recordExamples = Object.values({
		...paramExamples,
		...isObjectLike(contentExamples) ? contentExamples : {}
	}).filter(filterUndefined);
	const fallbackExample = recordExamples.length === 0 && "example" in parameter && parameter.example !== void 0 ? [parameter.example] : [];
	const arrayExamples = (schemaExamples ?? fallbackExample).filter(filterUndefined);
	return [...recordExamples, ...arrayExamples];
};
//#endregion
export { getParameterExamples };

//# sourceMappingURL=get-parameter-examples.js.map