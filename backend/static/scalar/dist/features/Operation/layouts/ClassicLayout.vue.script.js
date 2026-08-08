import { useLocalization } from "../../localization/use-localization.js";
import Badge_default from "../../../components/Badge/Badge.vue.js";
import SectionAccordion_default from "../../../components/Section/SectionAccordion.vue.js";
import Anchor_default from "../../../components/Anchor/Anchor.vue.js";
import { getXKeysFromObject } from "../../specification-extension/helpers.js";
import SpecificationExtension_default from "../../specification-extension/SpecificationExtension.vue.js";
import LinkList_default from "../../../components/LinkList/LinkList.vue.js";
import ExternalDocs_default from "../../external-docs/ExternalDocs.vue.js";
import { REQUEST_BODY_COMPOSITION_INDEX_SYMBOL } from "../request-body-composition-index.js";
import HttpMethod_default from "../../../components/HttpMethod/HttpMethod.vue.js";
import { getRequiredScopeGroups } from "../helpers/get-required-security.js";
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
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createSlots, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, provide, ref, toDisplayString, unref, withCtx, withModifiers } from "vue";
import { useClipboard } from "@scalar/use-hooks/useClipboard";
import { ScalarIconCopy, ScalarIconPlay, ScalarIconWebhooksLogo } from "@scalar/icons";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { ScalarIconButton } from "@scalar/components/icon-button";
import { CodeExample } from "@scalar/blocks/code-example";
import { ScalarErrorBoundary } from "@scalar/components/error-boundary";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/features/Operation/layouts/ClassicLayout.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "operation-title" };
var _hoisted_2 = { class: "operation-details" };
var _hoisted_3 = { class: "endpoint-label" };
var _hoisted_4 = { class: "endpoint-label-path" };
var _hoisted_5 = { class: "endpoint-label-name" };
var _hoisted_6 = {
	key: 1,
	class: "font-code text-sm"
};
var _hoisted_7 = { class: "endpoint-content" };
var _hoisted_8 = { class: "operation-details-card" };
var _hoisted_9 = {
	key: 0,
	class: "operation-details-card-item"
};
var _hoisted_10 = {
	key: 1,
	class: "operation-details-card-item"
};
var _hoisted_11 = { class: "operation-details-card-item" };
var _hoisted_12 = { class: "operation-details-card-item" };
var _hoisted_13 = {
	key: 2,
	class: "operation-details-card-item"
};
var ClassicLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ClassicLayout",
	props: {
		id: {},
		method: {},
		options: {},
		path: {},
		clientOptions: {},
		isCollapsed: { type: Boolean },
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
		const operationExtensions = computed(() => getXKeysFromObject(__props.operation));
		/** Whether the operation requires any OAuth scopes, used to skip the empty card item. */
		const hasRequiredScopes = computed(() => getRequiredScopeGroups(__props.requiredSecurity).length > 0);
		/** Track the selected request body content type so the code sample stays in sync */
		const selectedRequestBodyContentType = ref();
		/**
		* The example key actually shown in the request snippet for this operation.
		*
		* The test-request button lives in the accordion header, outside `OperationCodeSample`, so it
		* cannot read the resolved key from the footer slot like the modern layout does. We mirror it here
		* so the button opens the client with the same example the snippet displays, even when this
		* operation does not share the document-wide selection.
		*/
		const resolvedExampleKey = ref("");
		/** Selected request body oneOf/anyOf variants; synced with schema dropdowns and code sample */
		const requestBodyCompositionSelection = ref({});
		const requestBodyCompositionSelectionForCodeSample = computed(() => ({ ...requestBodyCompositionSelection.value }));
		const requestBodyCompositionSelectionKey = computed(() => JSON.stringify(requestBodyCompositionSelectionForCodeSample.value));
		provide(REQUEST_BODY_COMPOSITION_INDEX_SYMBOL, requestBodyCompositionSelection);
		const { copyToClipboard } = useClipboard();
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(SectionAccordion_default), {
				id: __props.id,
				"aria-label": operationTitle.value,
				class: "reference-endpoint",
				modelValue: !__props.isCollapsed,
				transparent: "",
				"onUpdate:modelValue": _cache[4] || (_cache[4] = (value) => __props.eventBus?.emit("toggle:nav-item", {
					id: __props.id,
					open: value
				}))
			}, createSlots({
				title: withCtx(() => [createElementVNode("div", _hoisted_1, [createElementVNode("div", _hoisted_2, [createVNode(unref(HttpMethod_default), {
					class: "endpoint-type",
					method: __props.method,
					short: ""
				}, null, 8, ["method"]), createVNode(unref(Anchor_default), {
					class: "endpoint-anchor",
					onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.id }))
				}, {
					default: withCtx(() => [createElementVNode("h3", _hoisted_3, [
						createElementVNode("div", _hoisted_4, [createVNode(OperationPath_default, {
							deprecated: unref(isOperationDeprecated)(__props.operation),
							path: __props.path
						}, null, 8, ["deprecated", "path"])]),
						createElementVNode("div", _hoisted_5, toDisplayString(operationTitle.value), 1),
						unref(getOperationStability)(__props.operation) ? (openBlock(), createBlock(unref(Badge_default), {
							key: 0,
							class: normalizeClass(["capitalize", unref(getOperationStabilityColor)(__props.operation)])
						}, {
							default: withCtx(() => [createTextVNode(toDisplayString(unref(getOperationStability)(__props.operation)), 1)]),
							_: 1
						}, 8, ["class"])) : createCommentVNode("", true),
						__props.isWebhook ? (openBlock(), createBlock(unref(Badge_default), {
							key: 1,
							class: "font-code text-green flex w-fit items-center justify-center gap-1"
						}, {
							default: withCtx(() => [createVNode(unref(ScalarIconWebhooksLogo), { weight: "bold" }), createTextVNode(" " + toDisplayString(unref(translate)("operation.webhook")), 1)]),
							_: 1
						})) : createCommentVNode("", true),
						createVNode(unref(XBadges_default), {
							badges: __props.operation["x-badges"],
							position: "before"
						}, null, 8, ["badges"])
					])]),
					_: 1
				})])])]),
				actions: withCtx(({ active }) => [
					createVNode(SecurityRequirementBadge_default, {
						hideLabel: "",
						requiredSecurity: __props.requiredSecurity
					}, null, 8, ["requiredSecurity"]),
					createVNode(unref(XBadges_default), {
						badges: __props.operation["x-badges"],
						position: "after"
					}, null, 8, ["badges"]),
					!__props.options.hideTestRequestButton ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [active && !__props.isWebhook ? (openBlock(), createBlock(unref(TestRequestButton_default), {
						key: 0,
						id: __props.id,
						eventBus: __props.eventBus,
						exampleName: resolvedExampleKey.value,
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
					])) : (openBlock(), createBlock(unref(ScalarIconPlay), {
						key: 1,
						class: "endpoint-try-hint size-4.5"
					}))], 64)) : createCommentVNode("", true),
					__props.options.showOperationId && __props.operation.operationId ? (openBlock(), createElementBlock("span", _hoisted_6, toDisplayString(__props.operation.operationId), 1)) : createCommentVNode("", true),
					createVNode(unref(ScalarIconButton), {
						class: "endpoint-copy p-0.5",
						icon: unref(ScalarIconCopy),
						label: unref(translate)("actions.copyEndpointUrl"),
						size: "xs",
						variant: "ghost",
						onClick: _cache[1] || (_cache[1] = withModifiers(($event) => unref(copyToClipboard)(__props.path), ["stop"]))
					}, null, 8, ["icon", "label"])
				]),
				default: withCtx(() => [createElementVNode("div", _hoisted_7, [
					createElementVNode("div", _hoisted_8, [
						Object.keys(operationExtensions.value).length > 0 ? (openBlock(), createElementBlock("div", _hoisted_9, [createVNode(SpecificationExtension_default, { value: operationExtensions.value }, null, 8, ["value"])])) : createCommentVNode("", true),
						hasRequiredScopes.value ? (openBlock(), createElementBlock("div", _hoisted_10, [createVNode(OperationScopes_default, { requiredSecurity: __props.requiredSecurity }, null, 8, ["requiredSecurity"])])) : createCommentVNode("", true),
						createElementVNode("div", _hoisted_11, [createVNode(OperationParameters_default, {
							selectedContentType: selectedRequestBodyContentType.value,
							"onUpdate:selectedContentType": _cache[2] || (_cache[2] = ($event) => selectedRequestBodyContentType.value = $event),
							document: __props.document,
							eventBus: __props.eventBus,
							options: __props.options,
							parameters: __props.operation.parameters,
							requestBody: unref(getResolvedRef)(__props.operation.requestBody)
						}, null, 8, [
							"selectedContentType",
							"document",
							"eventBus",
							"options",
							"parameters",
							"requestBody"
						])]),
						createElementVNode("div", _hoisted_12, [createVNode(OperationResponses_default, {
							document: __props.document,
							eventBus: __props.eventBus,
							options: __props.options,
							responses: __props.operation.responses
						}, null, 8, [
							"document",
							"eventBus",
							"options",
							"responses"
						])]),
						__props.operation?.callbacks ? (openBlock(), createElementBlock("div", _hoisted_13, [createVNode(Callbacks_default, {
							callbacks: __props.operation.callbacks,
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
						])])) : createCommentVNode("", true)
					]),
					__props.operation.responses ? (openBlock(), createBlock(unref(ExampleResponses_default), {
						key: 0,
						class: "operation-example-card",
						eventBus: __props.eventBus,
						responses: __props.operation.responses,
						selectedExample: __props.selectedExample
					}, null, 8, [
						"eventBus",
						"responses",
						"selectedExample"
					])) : createCommentVNode("", true),
					createElementVNode("div", null, [__props.operation.externalDocs ? (openBlock(), createBlock(unref(LinkList_default), { key: 0 }, {
						default: withCtx(() => [createVNode(unref(ExternalDocs_default), { value: __props.operation.externalDocs }, null, 8, ["value"])]),
						_: 1
					})) : createCommentVNode("", true), createVNode(unref(ScalarErrorBoundary), null, {
						default: withCtx(() => [(openBlock(), createBlock(unref(CodeExample), {
							key: requestBodyCompositionSelectionKey.value,
							class: "operation-example-card",
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
							selectedServer: __props.selectedServer,
							"onUpdate:exampleKey": _cache[3] || (_cache[3] = ($event) => resolvedExampleKey.value = $event)
						}, null, 8, [
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
					})])
				])]),
				_: 2
			}, [__props.operation.description ? {
				name: "description",
				fn: withCtx(() => [createVNode(unref(ScalarMarkdown), {
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
				key: "0"
			} : void 0]), 1032, [
				"id",
				"aria-label",
				"modelValue"
			]);
		};
	}
});
//#endregion
export { ClassicLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ClassicLayout.vue.script.js.map