import { getAsyncApiModelSchema } from "../../../helpers/get-async-api-model-schema.js";
import { isIntroductionEntry } from "../../../helpers/is-introduction-entry.js";
import { extractBodyDescriptions, extractBodyFieldNames, extractParameterDescriptions, extractParameterNames, extractSchemaDescriptions, extractSchemaFieldNames } from "../../../helpers/openapi.js";
import { DEFAULT_MODELS_SECTION_LABEL } from "@scalar/types/api-reference";
import { combineParams } from "@scalar/workspace-store/request-example";
import { isAsyncApiDocument, isOpenApiDocument } from "@scalar/workspace-store/schemas/type-guards";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { getPathItemOperation, getResolvedPathItem } from "@scalar/workspace-store/helpers/for-each-path-item-operation";
//#region src/features/Search/helpers/create-search-index.ts
/**
* Resolves a schema from `components.schemas` for either document type.
*
* OpenAPI and AsyncAPI keep reusable schemas in the same place, so model search entries can read
* property names and descriptions from both. AsyncAPI entries need extra handling (ref siblings,
* multi-format wrappers, boolean schemas), which lives in {@link getAsyncApiModelSchema}.
*/
function getModelSchema(document, name) {
	if (isOpenApiDocument(document)) return getResolvedRef(document.components?.schemas?.[name]);
	if (isAsyncApiDocument(document)) return getAsyncApiModelSchema(document, name);
}
function responseExampleValueToString(value) {
	if (typeof value === "string") return value;
	try {
		return JSON.stringify(value);
	} catch (_error) {
		return "";
	}
}
function mediaTypeExamplesToStrings(mediaType) {
	const examplesFromNamedMap = Object.values(mediaType.examples ?? {}).flatMap((example) => {
		const resolvedExample = getResolvedRef(example);
		if (!resolvedExample || !("value" in resolvedExample)) return [];
		return responseExampleValueToString(resolvedExample.value);
	}).filter((value) => value.length > 0);
	const mediaTypeExample = "example" in mediaType && mediaType.example !== void 0 ? responseExampleValueToString(mediaType.example) : "";
	return mediaTypeExample ? [mediaTypeExample, ...examplesFromNamedMap] : examplesFromNamedMap;
}
function extractResponseExamples(responses) {
	if (!responses) return [];
	return Object.values(responses).flatMap((response) => {
		const resolvedResponse = getResolvedRef(response);
		if (!resolvedResponse?.content) return [];
		return Object.values(resolvedResponse.content).flatMap((mediaType) => {
			const resolvedMediaType = getResolvedRef(mediaType);
			if (!resolvedMediaType) return [];
			return mediaTypeExamplesToStrings(resolvedMediaType);
		});
	}).filter((value) => value.length > 0);
}
var DEFAULT_SEARCH_INDEX_LABELS = {
	heading: "Heading",
	tagGroup: "Tag Group",
	webhook: "Webhook",
	webhooks: "Webhooks",
	introduction: "Introduction"
};
/**
* Create a search index from a list of entries.
*/
function createSearchIndex(document, options) {
	const index = [];
	const modelsSectionTitle = options?.modelsSectionLabel ?? DEFAULT_MODELS_SECTION_LABEL;
	const labels = options?.labels ?? DEFAULT_SEARCH_INDEX_LABELS;
	/**
	* Recursively processes entries and their children to build the search index.
	*/
	function processEntries(entriesToProcess) {
		entriesToProcess.forEach((entry) => {
			addEntryToIndex(entry, index, document, modelsSectionTitle, labels);
			if ("children" in entry && entry.children) processEntries(entry.children);
		});
	}
	processEntries(document?.["x-scalar-navigation"]?.children ?? []);
	return index;
}
/**
* Adds a single entry to the search index, handling all entry types recursively.
*
* AsyncAPI documents contribute heading, tag, and model entries here. Their
* channels, operations, and messages are not indexed yet.
*/
function addEntryToIndex(entry, index, document, modelsSectionTitle, labels) {
	const openApiDocument = isOpenApiDocument(document) ? document : void 0;
	if (entry.type === "operation") {
		const pathItem = getResolvedPathItem(openApiDocument?.paths?.[entry.path]);
		const operation = getResolvedRef(getPathItemOperation(openApiDocument?.paths?.[entry.path], entry.method)) ?? {};
		const operationWithPathParams = {
			...operation,
			parameters: combineParams(pathItem?.parameters, operation.parameters)
		};
		const parameters = extractParameterNames(operationWithPathParams.parameters ?? []);
		const parameterDescriptions = extractParameterDescriptions(operationWithPathParams.parameters ?? []);
		const body = extractBodyFieldNames(operationWithPathParams);
		const bodyDescriptions = extractBodyDescriptions(operationWithPathParams);
		const responseExamples = extractResponseExamples(operationWithPathParams.responses);
		index.push({
			type: "operation",
			title: entry.title,
			id: entry.id,
			description: operationWithPathParams.description || "",
			method: entry.method,
			path: entry.path,
			body,
			bodyDescriptions,
			parameters,
			parameterDescriptions,
			responseExamples,
			operationId: operationWithPathParams.operationId,
			entry
		});
		return;
	}
	if (entry.type === "webhook") {
		const webhook = getResolvedRef(getPathItemOperation(openApiDocument?.webhooks?.[entry.name], entry.method)) ?? {};
		const webhookDescription = webhook.description || "";
		index.push({
			id: entry.id,
			type: "webhook",
			title: entry.title,
			description: "Webhook",
			method: entry.method,
			body: "",
			bodyDescriptions: webhookDescription ? [webhookDescription] : [],
			operationId: webhook.operationId,
			entry
		});
		return;
	}
	if (entry.type === "model") {
		const schema = getModelSchema(document, entry.name);
		const schemaDescription = schema?.description ?? "";
		const propertyNames = extractSchemaFieldNames(schema);
		const propertyDescriptions = extractSchemaDescriptions(schema);
		index.push({
			type: "model",
			title: entry.title,
			description: modelsSectionTitle,
			id: entry.id,
			body: propertyNames,
			bodyDescriptions: schemaDescription ? [schemaDescription, ...propertyDescriptions] : propertyDescriptions,
			entry
		});
		return;
	}
	if (entry.type === "models") {
		index.push({
			id: entry.id,
			type: "heading",
			title: modelsSectionTitle,
			description: labels.heading,
			body: "",
			entry
		});
		return;
	}
	if (entry.type === "tag" && entry.isWebhooks === true) {
		index.push({
			id: entry.id,
			type: "heading",
			title: labels.webhooks,
			description: labels.heading,
			body: "",
			entry
		});
		return;
	}
	if (entry.type === "tag" && entry.isGroup === false) {
		index.push({
			id: entry.id,
			title: entry.title,
			description: entry.description || "",
			type: "tag",
			body: "",
			entry
		});
		return;
	}
	if (entry.type === "tag" && entry.isGroup === true) {
		index.push({
			id: entry.id,
			title: entry.title,
			description: labels.tagGroup,
			type: "tag",
			body: "",
			entry
		});
		return;
	}
	if (entry.type === "text") {
		index.push({
			id: entry.id,
			type: "heading",
			title: isIntroductionEntry(entry) ? labels.introduction : entry.title ?? "",
			description: labels.heading,
			body: "",
			entry
		});
		return;
	}
}
//#endregion
export { createSearchIndex };

//# sourceMappingURL=create-search-index.js.map