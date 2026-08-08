import { compositions } from "./schema-composition.js";
import { resolve } from "@scalar/workspace-store/resolve";
//#region src/components/Content/Schema/helpers/optimize-value-for-display.ts
/**
* Shallow-merges schema-like objects, but unions `properties` and `required`
* instead of letting a later object's `properties`/`required` completely
* overwrite an earlier one's. Without this, a variant schema (e.g. a `oneOf`
* branch) that declares its own `properties`/`required` would wipe out the
* shared base fields factored out at the root (or via a sibling `allOf`)
* instead of being combined with them (#9657).
*/
function mergeSchemaProperties(...objects) {
	const merged = {};
	const properties = {};
	const required = /* @__PURE__ */ new Set();
	let hasProperties = false;
	let hasRequired = false;
	for (const object of objects) {
		if (!object) continue;
		for (const [key, val] of Object.entries(object)) {
			if (key === "properties" && val && typeof val === "object") {
				Object.assign(properties, val);
				hasProperties = true;
				continue;
			}
			if (key === "required" && Array.isArray(val)) {
				for (const name of val) required.add(name);
				hasRequired = true;
				continue;
			}
			merged[key] = val;
		}
	}
	if (hasProperties) merged.properties = properties;
	if (hasRequired) merged.required = [...required];
	return merged;
}
/**
* Optimize the value by removing nulls from compositions and merging root properties.
*
* TODO: figure out what this does
*/
function optimizeValueForDisplay(value) {
	if (!value || typeof value !== "object") return value;
	const composition = compositions.find((keyword) => keyword in value && keyword !== "not");
	if (!composition) return { ...value };
	const schemas = value[composition];
	if (!Array.isArray(schemas)) return { ...value };
	const { [composition]: _, nullable: originalNullable, ...rootProperties } = value;
	const hasRootProperties = Object.keys(rootProperties).length > 0;
	const { filteredSchemas, hasNullSchema } = schemas.reduce((acc, _schema) => {
		const schema = resolve.schema(_schema);
		if ("type" in schema && schema.type === "null") acc.hasNullSchema = true;
		else acc.filteredSchemas.push(schema);
		return acc;
	}, {
		filteredSchemas: [],
		hasNullSchema: false
	});
	const shouldBeNullable = hasNullSchema || originalNullable === true;
	if (filteredSchemas.length === 1) {
		const mergedSchema = mergeSchemaProperties(filteredSchemas[0], rootProperties);
		if (shouldBeNullable) mergedSchema.nullable = true;
		return mergedSchema;
	}
	if ((composition === "oneOf" || composition === "anyOf") && (hasRootProperties || filteredSchemas.some((schema) => schema.allOf))) {
		const mergedSchemas = filteredSchemas.map((_schema) => {
			const schema = resolve.schema(_schema);
			if (schema.allOf?.length === 1) {
				const { allOf, ...otherProps } = schema;
				return mergeSchemaProperties(rootProperties, otherProps, resolve.schema(allOf[0]));
			}
			return mergeSchemaProperties(rootProperties, schema);
		});
		const result = { [composition]: mergedSchemas };
		if (typeof value.description === "string") result.description = value.description;
		if (shouldBeNullable) result.nullable = true;
		return result;
	}
	if (filteredSchemas.length !== schemas.length) {
		const result = {
			...value,
			[composition]: filteredSchemas
		};
		if (shouldBeNullable) result.nullable = true;
		return result;
	}
	return { ...value };
}
//#endregion
export { optimizeValueForDisplay };

//# sourceMappingURL=optimize-value-for-display.js.map