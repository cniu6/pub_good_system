import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { resolve } from "@scalar/workspace-store/resolve";
import { isArraySchema } from "@scalar/workspace-store/schemas/v3.1/strict/type-guards";
import { objectKeys } from "@scalar/helpers/object/object-keys";
//#region src/components/Content/Schema/helpers/merge-all-of-schemas.ts
/**
* Schema keywords whose value should reflect the *last* occurrence when merging
* `allOf` members. Most keywords keep the first occurrence, but for human-facing
* annotations a later subschema is expected to override an earlier one — matching
* OpenAPI/JSON Schema tooling like Swagger UI.
*/
var LAST_WINS_KEYS = /* @__PURE__ */ new Set(["description", "title"]);
/**
* Merges multiple OpenAPI schema objects into a single schema object.
* Handles nested allOf compositions and merges properties recursively.
*
* @param schemas - Array of OpenAPI schema objects to merge
* @param rootSchema - Optional root schema to merge with the result
* @param seenRefs - `$ref` strings already being merged higher in the call stack
* @returns Merged schema object
*/
var mergeAllOfSchemas = (schemas, rootSchema, seenRefs = /* @__PURE__ */ new Set()) => {
	if (!schemas?.allOf?.length || !Array.isArray(schemas.allOf)) return rootSchema || {};
	const result = {};
	const { allOf: _, ...baseSchema } = schemas;
	for (const _schema of schemas.allOf) {
		if (!_schema || typeof _schema !== "object") continue;
		const schema = resolve.schema(_schema);
		if (schema.allOf) {
			mergeSchemaIntoResult(result, mergeAllOfSchemas(schema, void 0, seenRefs), false, seenRefs);
			continue;
		}
		mergeSchemaIntoResult(result, schema, false, seenRefs);
	}
	if (Object.keys(baseSchema).length > 0) mergeSchemaIntoResult(result, baseSchema, true, seenRefs);
	if (rootSchema && typeof rootSchema === "object") if (rootSchema.allOf) mergeSchemaIntoResult(result, mergeAllOfSchemas(rootSchema, void 0, seenRefs), true, seenRefs);
	else mergeSchemaIntoResult(result, rootSchema, true, seenRefs);
	return result;
};
/**
* Efficiently merges a source schema into a target result object.
* Handles all schema merging logic in a single optimized function.
*
* @param result - The target schema object to merge into
* @param schema - The source schema object to merge from
* @param override - Whether to override existing properties (default: false)
*/
var mergeSchemaIntoResult = (result, schema, override = false, seenRefs = /* @__PURE__ */ new Set()) => {
	const schemaKeys = objectKeys(schema);
	if (schemaKeys.length === 0) return;
	for (const key of schemaKeys) {
		const value = getResolvedRef(schema[key]);
		if (value === void 0) continue;
		if (key === "required") {
			if (Array.isArray(value) && value.length > 0) if (result.required?.length) result.required = [.../* @__PURE__ */ new Set([...result.required, ...value])];
			else result.required = value.slice();
		} else if (key === "properties") {
			if (value && typeof value === "object") {
				if (!result.properties) result.properties = {};
				mergePropertiesIntoResult(result.properties, value, seenRefs);
			}
		} else if (key === "items") {
			const items = resolve.schema(value);
			if (items) {
				if (isArraySchema(schema)) {
					if (!result.items) result.items = {};
					if (items.allOf) {
						const mergedItems = mergeAllOfSchemas(items, void 0, seenRefs);
						Object.assign(result.items, mergedItems);
					} else mergeItemsIntoResult(getResolvedRef(result.items), items, seenRefs);
				} else if (items.allOf) {
					const mergedItems = mergeAllOfSchemas(items, void 0, seenRefs);
					if ("properties" in mergedItems) {
						if (!("properties" in result)) result.properties = {};
						"properties" in result && mergePropertiesIntoResult(result.properties, mergedItems.properties, seenRefs);
					}
				} else if (!("items" in result)) result.items = items;
			}
		} else if (key === "enum") {
			if (Array.isArray(value) && value.length > 0) result.enum = [.../* @__PURE__ */ new Set([...result.enum || [], ...value])];
		} else if (key === "oneOf" || key === "anyOf") {
			if (Array.isArray(value) && value.length > 0 && (override || result[key] === void 0)) result[key] = value;
		} else if (key === "allOf") continue;
		else if (override || LAST_WINS_KEYS.has(key) || result[key] === void 0) result[key] = value;
	}
};
/**
* Efficiently merges properties into a result object without creating new objects.
*/
var mergePropertiesIntoResult = (result, properties, seenRefs = /* @__PURE__ */ new Set()) => {
	const propertyKeys = Object.keys(properties ?? {});
	if (!properties || !result || propertyKeys.length === 0) return;
	for (const key of propertyKeys) {
		const schema = resolve.schema(properties[key]);
		if (!schema) {
			delete result[key];
			continue;
		}
		if (typeof schema !== "object") {
			result[key] = schema;
			continue;
		}
		if (!result[key]) {
			const rawProperty = properties[key];
			const newSchemaRef = rawProperty?.$ref;
			if (rawProperty && typeof newSchemaRef === "string" && seenRefs.has(newSchemaRef)) {
				result[key] = rawProperty;
				continue;
			}
			const nextNewSeenRefs = typeof newSchemaRef === "string" ? new Set(seenRefs).add(newSchemaRef) : seenRefs;
			if (schema.allOf) result[key] = mergeAllOfSchemas(schema, void 0, nextNewSeenRefs);
			else if (isArraySchema(schema) && resolve.schema(schema.items)?.allOf) result[key] = {
				...schema,
				items: mergeAllOfSchemas(resolve.schema(schema.items), void 0, seenRefs)
			};
			else if (properties[key]) result[key] = properties[key];
			continue;
		}
		const existing = resolve.schema(result[key]);
		const schemaRef = schema.$ref;
		if (typeof schemaRef === "string" && seenRefs.has(schemaRef)) {
			result[key] = existing;
			continue;
		}
		const nextSeenRefs = typeof schemaRef === "string" ? new Set(seenRefs).add(schemaRef) : seenRefs;
		if (schema.allOf) result[key] = mergeAllOfSchemas({ allOf: [existing, ...schema.allOf] }, void 0, nextSeenRefs);
		else if (isArraySchema(schema) && isArraySchema(existing) && schema.items) {
			const existingItems = resolve.schema(existing.items);
			result[key] = {
				...existing,
				type: "array",
				items: existingItems ? mergeItems(existingItems, resolve.schema(schema.items), nextSeenRefs) : resolve.schema(schema.items)
			};
		} else if ("properties" in existing && "properties" in schema) {
			const merged = {
				...existing,
				...schema
			};
			merged.properties = { ...existing.properties };
			mergePropertiesIntoResult(merged.properties, schema.properties, nextSeenRefs);
			result[key] = merged;
		} else result[key] = {
			...schema,
			...existing
		};
	}
};
/**
* Efficiently merges array items into a result object.
*/
var mergeItemsIntoResult = (result, items, seenRefs = /* @__PURE__ */ new Set()) => {
	if (items.allOf || result.allOf) {
		const allOfSchemas = [];
		if (result.allOf) for (const schema of result.allOf) allOfSchemas.push(resolve.schema(schema));
		else allOfSchemas.push(result);
		if (items.allOf) for (const schema of items.allOf) allOfSchemas.push(resolve.schema(schema));
		else allOfSchemas.push(items);
		const merged = mergeAllOfSchemas({ allOf: allOfSchemas }, void 0, seenRefs);
		Object.assign(result, merged);
		return;
	}
	Object.assign(result, items);
	if ("properties" in result && "properties" in items) mergePropertiesIntoResult(result.properties, items.properties, seenRefs);
};
/**
* Helper function for merging items that returns a new object.
*/
var mergeItems = (existing, incoming, seenRefs = /* @__PURE__ */ new Set()) => {
	const incomingRef = incoming.$ref;
	if (typeof incomingRef === "string") {
		if (seenRefs.has(incomingRef)) return existing;
		return mergeItemsInner(existing, incoming, new Set(seenRefs).add(incomingRef));
	}
	return mergeItemsInner(existing, incoming, seenRefs);
};
var mergeItemsInner = (existing, incoming, seenRefs = /* @__PURE__ */ new Set()) => {
	if (existing.allOf || incoming.allOf) {
		const allOfSchemas = [];
		if (existing.allOf) for (const schema of existing.allOf) allOfSchemas.push(resolve.schema(schema));
		else allOfSchemas.push(existing);
		if (incoming.allOf) for (const schema of incoming.allOf) allOfSchemas.push(resolve.schema(schema));
		else allOfSchemas.push(incoming);
		return mergeAllOfSchemas({ allOf: allOfSchemas }, void 0, seenRefs);
	}
	const merged = {
		...existing,
		...incoming
	};
	if ("properties" in existing && "properties" in incoming) {
		merged.properties = { ...existing.properties };
		mergePropertiesIntoResult(merged.properties, incoming.properties, seenRefs);
	}
	return merged;
};
//#endregion
export { mergeAllOfSchemas };

//# sourceMappingURL=merge-all-of-schemas.js.map