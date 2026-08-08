import { scrollTargetId } from "../../../helpers/lazy-bus.js";
import { getRefName } from "../../../components/Content/Schema/helpers/get-ref-name.js";
import SchemaProperty_default from "../../../components/Content/Schema/SchemaProperty.vue.js";
import ContentTypeSelect_default from "./ContentTypeSelect.vue.js";
import Headers_default from "./Headers.vue.js";
import { getParameterExamples } from "./helpers/get-parameter-examples.js";
import { computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, ref, resolveDynamicComponent, unref, withCtx } from "vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ScalarIconCaretRight } from "@scalar/icons";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { ScalarMarkdown, ScalarMarkdownSummary } from "@scalar/components/markdown";
import { ScalarWrappingText } from "@scalar/components/wrapping-text";
//#region src/features/Operation/components/ParameterListItem.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "parameter-item group/parameter-item" };
var _hoisted_2 = { class: "parameter-item-name min-w-0" };
var _hoisted_3 = {
	key: 1,
	class: "flex-1"
};
var ParameterListItem_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ParameterListItem",
	props: {
		parameter: {},
		name: {},
		breadcrumb: {},
		eventBus: {},
		collapsableItems: { type: Boolean },
		document: {},
		options: {}
	},
	setup(__props) {
		/** Whether the markdown summary is being truncated */
		const truncated = ref(false);
		/** Responses and params may both have a schema */
		const schema = computed(() => "schema" in __props.parameter && __props.parameter.schema ? getResolvedRef(__props.parameter.schema) : null);
		/** Response and params may both have content */
		const content = computed(() => {
			if (!("content" in __props.parameter) || !__props.parameter.content) return null;
			if (Object.keys(__props.parameter.content).length === 0) return null;
			return __props.parameter.content;
		});
		const selectedContentType = ref(Object.keys(content.value || {})[0] ?? "");
		/** Response headers */
		const headers = computed(() => "headers" in __props.parameter && __props.parameter.headers ? __props.parameter.headers : null);
		/** Raw schema (possibly with $ref) for the selected content type or param. */
		const baseSchema = computed(() => content.value ? content.value?.[selectedContentType.value]?.schema : "schema" in __props.parameter && __props.parameter.schema ? __props.parameter.schema : null);
		/** When the schema is a $ref, preserve its name so the UI can show the ref name instead of just the type. */
		const schemaModelName = computed(() => {
			const raw = baseSchema.value;
			if (!raw) return null;
			if ("$ref" in raw) return getRefName(raw.$ref);
			return null;
		});
		/** Computed value from the combined schema param and content param */
		const value = computed(() => {
			const base = baseSchema.value;
			const resolvedBase = content.value ? getResolvedRef(base) : schema.value;
			const deprecated = "deprecated" in __props.parameter ? __props.parameter.deprecated : schema.value?.deprecated;
			/** Combine param/content/schema examples while ignoring undefined values. */
			const examples = getParameterExamples({
				parameter: __props.parameter,
				schemaExamples: schema.value?.examples,
				contentExamples: content.value?.[selectedContentType.value]?.examples
			});
			return {
				...resolvedBase,
				deprecated,
				examples
			};
		});
		/**
		* Determines whether this parameter item should be rendered as a collapsible disclosure.
		* Only collapses when collapsableItems is enabled and the parameter has additional
		* content to display (content types, headers, or schema details).
		*/
		const shouldCollapse = computed(() => Boolean(content.value || headers.value || schema.value || truncated.value));
		/**
		* The breadcrumb passed to the schema. Collapsible items (responses) render their
		* schema without a name to avoid a duplicate heading, so we push the item name
		* (e.g. the status code) onto the breadcrumb here to keep property anchors unique.
		*/
		const schemaBreadcrumb = computed(() => __props.collapsableItems && __props.breadcrumb && __props.name ? [...__props.breadcrumb, __props.name] : __props.breadcrumb);
		/**
		* Whether a deep link points at a property inside this collapsed item. When it
		* does, the disclosure opens on mount so a fresh navigation can render the target
		* and scroll it into view (mirrors how collapsible schema disclosures behave).
		*/
		const isOnScrollTargetPath = computed(() => {
			const path = schemaBreadcrumb.value?.join(".");
			if (!path) return false;
			const target = scrollTargetId.value;
			return target === path || target.startsWith(`${path}.`);
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("li", _hoisted_1, [createVNode(unref(Disclosure), { defaultOpen: isOnScrollTargetPath.value }, {
				default: withCtx(({ open }) => [
					__props.collapsableItems ? (openBlock(), createBlock(resolveDynamicComponent(shouldCollapse.value ? unref(DisclosureButton) : "div"), {
						key: 0,
						class: normalizeClass(["parameter-item-trigger", { "parameter-item-trigger-open": open }])
					}, {
						default: withCtx(() => [createElementVNode("div", _hoisted_2, [shouldCollapse.value ? (openBlock(), createBlock(unref(ScalarIconCaretRight), {
							key: 0,
							class: normalizeClass(["parameter-item-icon size-3 transition-transform duration-100", { "rotate-90": open }]),
							weight: "bold"
						}, null, 8, ["class"])) : createCommentVNode("", true), createElementVNode("div", null, [createVNode(unref(ScalarWrappingText), {
							preset: "property",
							text: __props.name
						}, null, 8, ["text"])])]), !open && __props.parameter.description ? (openBlock(), createBlock(unref(ScalarMarkdownSummary), {
							key: 0,
							truncated: truncated.value,
							"onUpdate:truncated": _cache[0] || (_cache[0] = ($event) => truncated.value = $event),
							class: "parameter-item-description-summary min-w-0 flex-1",
							controlled: "",
							value: __props.parameter.description
						}, null, 8, ["truncated", "value"])) : (openBlock(), createElementBlock("div", _hoisted_3))]),
						_: 2
					}, 1032, ["class"])) : createCommentVNode("", true),
					createVNode(unref(DisclosurePanel), {
						class: "parameter-item-container parameter-item-container-markdown",
						static: !__props.collapsableItems
					}, {
						default: withCtx(() => [
							__props.collapsableItems && __props.parameter.description ? (openBlock(), createBlock(unref(ScalarMarkdown), {
								key: 0,
								class: "parameter-item-description",
								value: __props.parameter.description
							}, null, 8, ["value"])) : createCommentVNode("", true),
							headers.value ? (openBlock(), createBlock(Headers_default, {
								key: 1,
								breadcrumb: __props.breadcrumb,
								document: __props.document,
								eventBus: __props.eventBus,
								expandAllSchemaProperties: __props.options.expandAllSchemaProperties,
								headers: headers.value,
								orderRequiredPropertiesFirst: __props.options.orderRequiredPropertiesFirst,
								orderSchemaPropertiesBy: __props.options.orderSchemaPropertiesBy
							}, null, 8, [
								"breadcrumb",
								"document",
								"eventBus",
								"expandAllSchemaProperties",
								"headers",
								"orderRequiredPropertiesFirst",
								"orderSchemaPropertiesBy"
							])) : createCommentVNode("", true),
							createVNode(SchemaProperty_default, {
								is: "div",
								breadcrumb: schemaBreadcrumb.value,
								compact: "",
								description: __props.collapsableItems ? "" : __props.parameter.description,
								eventBus: __props.eventBus,
								hideWriteOnly: true,
								modelName: schemaModelName.value,
								name: __props.collapsableItems ? "" : __props.name,
								noncollapsible: true,
								options: {
									hideWriteOnly: true,
									orderRequiredPropertiesFirst: __props.options.orderRequiredPropertiesFirst,
									orderSchemaPropertiesBy: __props.options.orderSchemaPropertiesBy,
									expandAllSchemaProperties: __props.options.expandAllSchemaProperties,
									document: __props.document
								},
								required: "required" in __props.parameter && __props.parameter.required,
								schema: value.value
							}, null, 8, [
								"breadcrumb",
								"description",
								"eventBus",
								"modelName",
								"name",
								"options",
								"required",
								"schema"
							])
						]),
						_: 1
					}, 8, ["static"]),
					shouldCollapse.value && content.value ? (openBlock(), createElementBlock("div", {
						key: 1,
						class: normalizeClass(["absolute top-[calc(10px+0.5lh)] right-0 z-0 flex -translate-y-1/2 items-center text-base", { "opacity-0 group-focus-within/parameter-item:opacity-100 group-hover/parameter-item:opacity-100": !open }])
					}, [_cache[2] || (_cache[2] = createElementVNode("div", { class: "from-b-1 absolute inset-y-0 -left-6 -z-1 w-8 bg-linear-to-l from-40% to-transparent" }, null, -1)), createVNode(ContentTypeSelect_default, {
						modelValue: selectedContentType.value,
						"onUpdate:modelValue": _cache[1] || (_cache[1] = ($event) => selectedContentType.value = $event),
						content: content.value
					}, null, 8, ["modelValue", "content"])], 2)) : createCommentVNode("", true)
				]),
				_: 1
			}, 8, ["defaultOpen"])]);
		};
	}
});
//#endregion
export { ParameterListItem_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ParameterListItem.vue.script.js.map