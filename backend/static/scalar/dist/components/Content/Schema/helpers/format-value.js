//#region src/components/Content/Schema/helpers/format-value.ts
/**
* Converts a value to a string that can be displayed in the UI.
*/
function formatValue(value) {
	if (Array.isArray(value)) return `[${value.map((item) => {
		if (typeof item === "string") return `"${item.toString().trim()}"`;
		if (typeof item === "object") return JSON.stringify(item);
		if (item === void 0) return "undefined";
		if (item === null) return "null";
		return item;
	}).join(", ")}]`;
	if (value === null) return "null";
	if (typeof value === "object") return JSON.stringify(value);
	if (value === void 0) return "undefined";
	if (typeof value === "string") return value.trim();
	return value.toString().trim();
}
//#endregion
export { formatValue };

//# sourceMappingURL=format-value.js.map