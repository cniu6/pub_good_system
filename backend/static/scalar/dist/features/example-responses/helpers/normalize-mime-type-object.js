import { normalizeMimeType } from "./normalize-mime-type.js";
//#region src/features/example-responses/helpers/normalize-mime-type-object.ts
/**
* Remove charset from content types
*
* Example: `application/json; charset=utf-8` -> `application/json`
*/
function normalizeMimeTypeObject(content) {
	if (!content) return content;
	const newContent = { ...content };
	Object.entries(newContent).forEach(([key, value]) => {
		const normalizedKey = normalizeMimeType(key);
		if (normalizedKey) newContent[normalizedKey] = value;
	});
	return newContent;
}
//#endregion
export { normalizeMimeTypeObject };

//# sourceMappingURL=normalize-mime-type-object.js.map