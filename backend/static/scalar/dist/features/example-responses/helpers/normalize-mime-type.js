//#region src/features/example-responses/helpers/normalize-mime-type.ts
function normalizeMimeType(contentType) {
	if (typeof contentType !== "string") return;
	return contentType.replace(/;.*$/, "").replace(/\/(?!.*vnd\.|fhir\+).*\+/, "/").trim();
}
//#endregion
export { normalizeMimeType };

//# sourceMappingURL=normalize-mime-type.js.map