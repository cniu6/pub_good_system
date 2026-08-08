import { useLocalization } from "../../localization/use-localization.js";
import ParameterList_default from "./ParameterList.vue.js";
import { flattenDeepObjectQueryParameter } from "../helpers/flatten-deep-object-query-parameter.js";
import { shouldIgnoreEntity } from "../helpers/should-ignore-entity.js";
import RequestBody_default from "./RequestBody.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createTextVNode, createVNode, defineComponent, mergeModels, openBlock, toDisplayString, unref, useModel, withCtx } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/features/Operation/components/OperationParameters.vue?vue&type=script&setup=true&lang.ts
/** Use a single loop to reduce parameters by type(in) */
var OperationParameters_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "OperationParameters",
	props: /*@__PURE__*/ mergeModels({
		breadcrumb: {},
		parameters: { default: () => [] },
		requestBody: {},
		eventBus: {},
		document: {},
		options: {}
	}, {
		"selectedContentType": {},
		"selectedContentTypeModifiers": {}
	}),
	emits: ["update:selectedContentType"],
	setup(__props) {
		const { translate } = useLocalization();
		/** Thread the selected request body content type up to the layout */
		const selectedContentType = useModel(__props, "selectedContentType");
		const splitParameters = computed(() => (__props.parameters ?? []).reduce((acc, p) => {
			const parameter = getResolvedRef(p);
			if (!shouldIgnoreEntity(parameter)) flattenDeepObjectQueryParameter(parameter).forEach((flattenedParameter) => {
				acc[flattenedParameter.in].push(flattenedParameter);
			});
			return acc;
		}, {
			cookie: [],
			header: [],
			path: [],
			query: []
		}));
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [
				createVNode(ParameterList_default, {
					breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "path"] : void 0,
					document: __props.document,
					eventBus: __props.eventBus,
					options: __props.options,
					parameters: splitParameters.value["path"]
				}, {
					title: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("operation.pathParameters")), 1)]),
					_: 1
				}, 8, [
					"breadcrumb",
					"document",
					"eventBus",
					"options",
					"parameters"
				]),
				createVNode(ParameterList_default, {
					breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "query"] : void 0,
					document: __props.document,
					eventBus: __props.eventBus,
					options: __props.options,
					parameters: splitParameters.value["query"]
				}, {
					title: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("operation.queryParameters")), 1)]),
					_: 1
				}, 8, [
					"breadcrumb",
					"document",
					"eventBus",
					"options",
					"parameters"
				]),
				createVNode(ParameterList_default, {
					breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "headers"] : void 0,
					document: __props.document,
					eventBus: __props.eventBus,
					options: __props.options,
					parameters: splitParameters.value["header"]
				}, {
					title: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("operation.headers")), 1)]),
					_: 1
				}, 8, [
					"breadcrumb",
					"document",
					"eventBus",
					"options",
					"parameters"
				]),
				createVNode(ParameterList_default, {
					breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "cookies"] : void 0,
					document: __props.document,
					eventBus: __props.eventBus,
					options: __props.options,
					parameters: splitParameters.value["cookie"]
				}, {
					title: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("operation.cookies")), 1)]),
					_: 1
				}, 8, [
					"breadcrumb",
					"document",
					"eventBus",
					"options",
					"parameters"
				]),
				__props.requestBody ? (openBlock(), createBlock(RequestBody_default, {
					key: 0,
					selectedContentType: selectedContentType.value,
					"onUpdate:selectedContentType": _cache[0] || (_cache[0] = ($event) => selectedContentType.value = $event),
					breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "body"] : void 0,
					document: __props.document,
					eventBus: __props.eventBus,
					options: __props.options,
					requestBody: __props.requestBody
				}, {
					title: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("operation.body")), 1)]),
					_: 1
				}, 8, [
					"selectedContentType",
					"breadcrumb",
					"document",
					"eventBus",
					"options",
					"requestBody"
				])) : createCommentVNode("", true)
			], 64);
		};
	}
});
//#endregion
export { OperationParameters_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=OperationParameters.vue.script.js.map