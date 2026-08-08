import { useLocalization } from "../../localization/use-localization.js";
import Badge_default from "../../../components/Badge/Badge.vue.js";
import Section_default from "../../../components/Section/Section.vue.js";
import SectionContent_default from "../../../components/Section/SectionContent.vue.js";
import SectionHeaderTag_default from "../../../components/Section/SectionHeaderTag.vue.js";
import Anchor_default from "../../../components/Anchor/Anchor.vue.js";
import { getXKeysFromObject } from "../../specification-extension/helpers.js";
import SpecificationExtension_default from "../../specification-extension/SpecificationExtension.vue.js";
import LinkList_default from "../../../components/LinkList/LinkList.vue.js";
import ExternalDocs_default from "../../external-docs/ExternalDocs.vue.js";
import { REQUEST_BODY_COMPOSITION_INDEX_SYMBOL } from "../request-body-composition-index.js";
import OperationScopes_default from "../components/OperationScopes.vue.js";
import OperationPath_default from "../../../components/OperationPath.vue.js";
import ExampleResponses_default from "../../example-responses/ExampleResponses.vue.js";
import OperationParameters_default from "../components/OperationParameters.vue.js";
import OperationResponses_default from "../components/OperationResponses.vue.js";
import Callbacks_default from "../components/callbacks/Callbacks.vue.js";
import SecurityRequirementBadge_default from "../components/SecurityRequirementBadge.vue.js";
import { getOperationStability, getOperationStabilityColor, isOperationDeprecated } from "../helpers/operation-stability.js";
import TestRequestButton_default from "../../test-request-button/TestRequestButton.vue.js";
import XBadges_default from "../../x-badges/XBadges.vue.js";
import AskAgentButton_default from "../../ask-agent-button/AskAgentButton.vue.js";
import { computed, createBlock, createCommentVNode, createElementVNode, createSlots, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, provide, ref, toDisplayString, unref, useId, withCtx } from "vue";
import { ScalarIconWebhooksLogo } from "@scalar/icons";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { CodeExample } from "@scalar/blocks/code-example";
import { ScalarErrorBoundary } from "@scalar/components/error-boundary";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/features/Operation/layouts/ModernLayout.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "flex flex-row justify-between gap-1" };
var _hoisted_2 = { class: "flex gap-1" };
var _hoisted_3 = { class: "flex gap-1" };
var _hoisted_4 = { class: "operation-layout" };
var _hoisted_5 = { class: "operation-auth" };
var _hoisted_6 = { class: "operation-description" };
var _hoisted_7 = { class: "operation-details" };
var _hoisted_8 = { class: "examples" };
var _hoisted_9 = { class: "flex" };
var ModernLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ModernLayout",
	props: {
		id: {},
		method: {},
		options: {},
		path: {},
		clientOptions: {},
		isWebhook: { type: Boolean },
		selectedClient: {},
		selectedExample: {},
		eventBus: {},
		operation: {},
		selectedServer: {},
		selectedSecuritySchemes: {},
		requiredSecurity: {},
		document: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		const operationTitle = computed(() => __props.operation.summary || __props.path || "");
		const labelId = useId();
		const operationExtensions = computed(() => getXKeysFromObject(__props.operation));
		/** Track the selected request body content type so the code sample stays in sync */
		const selectedRequestBodyContentType = ref();
		/** Selected request body oneOf/anyOf variants; synced with schema dropdowns and code sample */
		const requestBodyCompositionSelection = ref({});
		const requestBodyCompositionSelectionForCodeSample = computed(() => ({ ...requestBodyCompositionSelection.value }));
		const requestBodyCompositionSelectionKey = computed(() => JSON.stringify(requestBodyCompositionSelectionForCodeSample.value));
		provide(REQUEST_BODY_COMPOSITION_INDEX_SYMBOL, requestBodyCompositionSelection);
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(Section_default), {
				id: __props.id,
				"aria-labelledby": unref(labelId),
				label: operationTitle.value,
				tabindex: "-1",
				onIntersecting: _cache[2] || (_cache[2] = () => __props.eventBus?.emit("intersecting:nav-item", { id: __props.id }))
			}, {
				default: withCtx(() => [createVNode(unref(SectionContent_default), null, {
					default: withCtx(() => [createElementVNode("div", _hoisted_1, [createElementVNode("div", _hoisted_2, [
						__props.options?.showOperationId && __props.operation.operationId ? (openBlock(), createBlock(unref(Badge_default), { key: 0 }, {
							default: withCtx(() => [createTextVNode(toDisplayString(__props.operation.operationId), 1)]),
							_: 1
						})) : createCommentVNode("", true),
						unref(getOperationStability)(__props.operation) ? (openBlock(), createBlock(unref(Badge_default), {
							key: 1,
							class: normalizeClass(["capitalize", unref(getOperationStabilityColor)(__props.operation)])
						}, {
							default: withCtx(() => [createTextVNode(toDisplayString(unref(getOperationStability)(__props.operation)), 1)]),
							_: 1
						}, 8, ["class"])) : createCommentVNode("", true),
						__props.isWebhook ? (openBlock(), createBlock(unref(Badge_default), {
							key: 2,
							class: "font-code text-green flex w-fit items-center justify-center gap-1"
						}, {
							default: withCtx(() => [createVNode(unref(ScalarIconWebhooksLogo), { weight: "bold" }), createTextVNode(" " + toDisplayString(unref(translate)("operation.webhook")), 1)]),
							_: 1
						})) : createCommentVNode("", true),
						createVNode(unref(XBadges_default), {
							badges: __props.operation["x-badges"],
							position: "before"
						}, null, 8, ["badges"])
					]), createElementVNode("div", _hoisted_3, [createVNode(unref(XBadges_default), {
						badges: __props.operation["x-badges"],
						position: "after"
					}, null, 8, ["badges"])])]), createElementVNode("div", _hoisted_4, [
						createElementVNode("div", { class: normalizeClass(["operation-title", unref(isOperationDeprecated)(__props.operation) && "deprecated"]) }, [createVNode(unref(Anchor_default), { onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.id })) }, {
							default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), {
								id: unref(labelId),
								level: 3
							}, {
								default: withCtx(() => [createTextVNode(toDisplayString(operationTitle.value), 1)]),
								_: 1
							}, 8, ["id"])]),
							_: 1
						})], 2),
						createElementVNode("div", _hoisted_5, [createVNode(SecurityRequirementBadge_default, { requiredSecurity: __props.requiredSecurity }, null, 8, ["requiredSecurity"])]),
						createElementVNode("div", _hoisted_6, [createVNode(SpecificationExtension_default, { value: operationExtensions.value }, null, 8, ["value"]), createVNode(unref(ScalarMarkdown), {
							anchorPrefix: __props.id,
							"aria-label": unref(translate)("common.description"),
							role: "group",
							transformType: "heading",
							value: __props.operation.description,
							withAnchors: "",
							withImages: ""
						}, null, 8, [
							"anchorPrefix",
							"aria-label",
							"value"
						])]),
						createElementVNode("div", _hoisted_7, [
							createVNode(OperationScopes_default, { requiredSecurity: __props.requiredSecurity }, null, 8, ["requiredSecurity"]),
							createVNode(OperationParameters_default, {
								selectedContentType: selectedRequestBodyContentType.value,
								"onUpdate:selectedContentType": _cache[1] || (_cache[1] = ($event) => selectedRequestBodyContentType.value = $event),
								breadcrumb: [__props.id],
								document: __props.document,
								eventBus: __props.eventBus,
								options: __props.options,
								parameters: __props.operation.parameters,
								requestBody: unref(getResolvedRef)(__props.operation.requestBody)
							}, null, 8, [
								"selectedContentType",
								"breadcrumb",
								"document",
								"eventBus",
								"options",
								"parameters",
								"requestBody"
							]),
							createVNode(OperationResponses_default, {
								breadcrumb: [__props.id],
								collapsableItems: !__props.options.expandAllResponses,
								document: __props.document,
								eventBus: __props.eventBus,
								options: __props.options,
								responses: __props.operation.responses
							}, null, 8, [
								"breadcrumb",
								"collapsableItems",
								"document",
								"eventBus",
								"options",
								"responses"
							]),
							createVNode(unref(ScalarErrorBoundary), null, {
								default: withCtx(() => [__props.operation.callbacks ? (openBlock(), createBlock(Callbacks_default, {
									key: 0,
									callbacks: __props.operation.callbacks,
									class: "mt-6",
									document: __props.document,
									eventBus: __props.eventBus,
									options: __props.options,
									path: __props.path
								}, null, 8, [
									"callbacks",
									"document",
									"eventBus",
									"options",
									"path"
								])) : createCommentVNode("", true)]),
								_: 1
							})
						]),
						createElementVNode("div", _hoisted_8, [
							__props.operation.externalDocs ? (openBlock(), createBlock(unref(LinkList_default), { key: 0 }, {
								default: withCtx(() => [createVNode(unref(ExternalDocs_default), { value: __props.operation.externalDocs }, null, 8, ["value"])]),
								_: 1
							})) : createCommentVNode("", true),
							createVNode(unref(ScalarErrorBoundary), null, {
								default: withCtx(() => [(openBlock(), createBlock(unref(CodeExample), {
									key: requestBodyCompositionSelectionKey.value,
									clientOptions: __props.clientOptions,
									eventBus: __props.eventBus,
									fallback: "",
									isWebhook: __props.isWebhook,
									method: __props.method,
									operation: __props.operation,
									path: __props.path,
									requestBodyCompositionSelection: requestBodyCompositionSelectionForCodeSample.value,
									securitySchemes: __props.selectedSecuritySchemes,
									selectedClient: __props.selectedClient,
									selectedContentType: selectedRequestBodyContentType.value,
									selectedExample: __props.selectedExample,
									selectedServer: __props.selectedServer
								}, createSlots({
									header: withCtx(() => [createVNode(OperationPath_default, {
										class: "font-code text-c-2 [&_em]:text-c-1 min-w-0 [&_em]:not-italic",
										deprecated: __props.operation?.deprecated,
										path: __props.path
									}, null, 8, ["deprecated", "path"])]),
									_: 2
								}, [!__props.isWebhook ? {
									name: "footer",
									fn: withCtx(({ exampleName }) => [createElementVNode("div", _hoisted_9, [createVNode(AskAgentButton_default), !__props.options.hideTestRequestButton ? (openBlock(), createBlock(unref(TestRequestButton_default), {
										key: 0,
										id: __props.id,
										eventBus: __props.eventBus,
										exampleName,
										method: __props.method,
										path: __props.path,
										requestBodyCompositionSelection: requestBodyCompositionSelectionForCodeSample.value
									}, null, 8, [
										"id",
										"eventBus",
										"exampleName",
										"method",
										"path",
										"requestBodyCompositionSelection"
									])) : createCommentVNode("", true)])]),
									key: "0"
								} : void 0]), 1032, [
									"clientOptions",
									"eventBus",
									"isWebhook",
									"method",
									"operation",
									"path",
									"requestBodyCompositionSelection",
									"securitySchemes",
									"selectedClient",
									"selectedContentType",
									"selectedExample",
									"selectedServer"
								]))]),
								_: 1
							}),
							createVNode(unref(ScalarErrorBoundary), null, {
								default: withCtx(() => [__props.operation.responses ? (openBlock(), createBlock(unref(ExampleResponses_default), {
									key: 0,
									eventBus: __props.eventBus,
									responses: __props.operation.responses,
									selectedExample: __props.selectedExample,
									style: { "margin-top": "12px" }
								}, null, 8, [
									"eventBus",
									"responses",
									"selectedExample"
								])) : createCommentVNode("", true)]),
								_: 1
							})
						])
					])]),
					_: 1
				})]),
				_: 1
			}, 8, [
				"id",
				"aria-labelledby",
				"label"
			]);
		};
	}
});
//#endregion
export { ModernLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ModernLayout.vue.script.js.map