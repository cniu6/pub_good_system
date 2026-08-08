import { useLocalization } from "../../localization/use-localization.js";
import { inferDiscriminatorMappingComposition } from "../../../components/Content/Schema/helpers/get-compositions-to-render.js";
import { isTypeObject } from "../../../components/Content/Schema/helpers/is-type-object.js";
import { getModelNameFromSchema } from "../../../components/Content/Schema/helpers/schema-name.js";
import { reduceNamesToObject, sortPropertyNames } from "../../../components/Content/Schema/helpers/sort-property-names.js";
import LinkButton_default from "../../../components/Content/Schema/LinkButton.vue.js";
import Schema_default from "../../../components/Content/Schema/Schema.vue.js";
import ContentTypeSelect_default from "./ContentTypeSelect.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, mergeModels, openBlock, renderSlot, toDisplayString, unref, useModel, withCtx } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/features/Operation/components/RequestBody.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["aria-label"];
var _hoisted_2 = { class: "request-body-header" };
var _hoisted_3 = { class: "request-body-title" };
var _hoisted_4 = {
	key: 0,
	class: "text-c-2 text-xs leading-none font-normal",
	"data-testid": "request-body-schema-name"
};
var _hoisted_5 = { class: "flex items-center gap-2" };
var _hoisted_6 = {
	key: 0,
	class: "request-body-required"
};
var _hoisted_7 = {
	key: 0,
	class: "request-body-description"
};
var _hoisted_8 = {
	key: 0,
	class: "request-body-schema"
};
var _hoisted_9 = {
	key: 1,
	class: "request-body-schema"
};
var MAX_VISIBLE_PROPERTIES = 12;
var RequestBody_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "RequestBody",
	props: /*@__PURE__*/ mergeModels({
		breadcrumb: {},
		requestBody: {},
		eventBus: {},
		document: {},
		options: {}
	}, {
		"selectedContentType": { default: "application/json" },
		"selectedContentTypeModifiers": {}
	}),
	emits: ["update:selectedContentType"],
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* The maximum number of properties to show in the request body schema.
		*/
		const availableContentTypes = computed(() => Object.keys(__props.requestBody?.content ?? {}));
		const selectedContentType = useModel(__props, "selectedContentType");
		if (__props.requestBody?.content) {
			if (availableContentTypes.value[0]) selectedContentType.value = availableContentTypes.value[0];
		}
		/** Raw schema (possibly with $ref) for the selected content type */
		const rawSchema = computed(() => __props.requestBody?.content?.[selectedContentType.value]?.schema);
		const schema = computed(() => getResolvedRef(rawSchema.value));
		/** When the schema is a $ref, preserve its name so the UI can show the ref name instead of just the type. */
		const modelLink = computed(() => (rawSchema.value && getModelNameFromSchema(rawSchema.value)) ?? null);
		/**
		* Splits schema properties into visible and collapsed sections when there are more than 12 properties.
		* Returns null for schemas with fewer properties or non-object schemas.
		*/
		const partitionedSchema = computed(() => {
			if (!schema.value || !isTypeObject(schema.value)) return null;
			if (inferDiscriminatorMappingComposition(schema.value, __props.document)) return null;
			const sortedNames = sortPropertyNames(schema.value, schema.value.discriminator, {
				hideReadOnly: true,
				orderSchemaPropertiesBy: __props.options.orderSchemaPropertiesBy,
				orderRequiredPropertiesFirst: __props.options.orderRequiredPropertiesFirst
			});
			if (sortedNames.length <= MAX_VISIBLE_PROPERTIES) return null;
			const { properties, ...schemaMetadata } = schema.value;
			if (!properties) return null;
			return {
				visibleProperties: {
					...schemaMetadata,
					properties: reduceNamesToObject(sortedNames.slice(0, MAX_VISIBLE_PROPERTIES), properties)
				},
				collapsedProperties: {
					...schemaMetadata,
					properties: reduceNamesToObject(sortedNames.slice(MAX_VISIBLE_PROPERTIES), properties)
				}
			};
		});
		/**
		* We don't want to render the request body if its completely empty
		* @example
		* {
		*   "content": {},
		* }
		*/
		const shouldRenderRequestBody = computed(() => Object.keys(__props.requestBody?.content ?? {}).length > 0 || __props.requestBody?.description || __props.requestBody?.required);
		return (_ctx, _cache) => {
			return __props.requestBody && shouldRenderRequestBody.value ? (openBlock(), createElementBlock("div", {
				key: 0,
				"aria-label": unref(translate)("operation.requestBody"),
				class: "request-body",
				role: "group"
			}, [createElementVNode("div", _hoisted_2, [
				createElementVNode("div", _hoisted_3, [renderSlot(_ctx.$slots, "title", {}, void 0, true), modelLink.value ? (openBlock(), createElementBlock("span", _hoisted_4, [_cache[2] || (_cache[2] = createElementVNode("span", { class: "text-c-3 mx-1.5" }, "·", -1)), __props.eventBus && modelLink.value.schemaKey ? (openBlock(), createBlock(LinkButton_default, {
					key: 0,
					onClick: _cache[0] || (_cache[0] = ($event) => __props.eventBus.emit("scroll-to:model-by-name", { name: modelLink.value.schemaKey }))
				}, {
					default: withCtx(() => [createTextVNode(toDisplayString(modelLink.value.label), 1)]),
					_: 1
				})) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(modelLink.value.label), 1)], 64))])) : createCommentVNode("", true)]),
				createElementVNode("div", _hoisted_5, [__props.requestBody.required ? (openBlock(), createElementBlock("div", _hoisted_6, toDisplayString(unref(translate)("common.required")), 1)) : createCommentVNode("", true), createVNode(ContentTypeSelect_default, {
					modelValue: selectedContentType.value,
					"onUpdate:modelValue": _cache[1] || (_cache[1] = ($event) => selectedContentType.value = $event),
					content: __props.requestBody.content
				}, null, 8, ["modelValue", "content"])]),
				__props.requestBody.description ? (openBlock(), createElementBlock("div", _hoisted_7, [createVNode(unref(ScalarMarkdown), { value: __props.requestBody.description }, null, 8, ["value"])])) : createCommentVNode("", true)
			]), partitionedSchema.value ? (openBlock(), createElementBlock("div", _hoisted_8, [createVNode(unref(Schema_default), {
				breadcrumb: __props.breadcrumb,
				compact: "",
				compositionPath: ["requestBody"],
				eventBus: __props.eventBus,
				name: unref(translate)("operation.requestBody"),
				noncollapsible: "",
				options: {
					hideReadOnly: true,
					orderRequiredPropertiesFirst: __props.options.orderRequiredPropertiesFirst,
					orderSchemaPropertiesBy: __props.options.orderSchemaPropertiesBy,
					expandAllSchemaProperties: __props.options.expandAllSchemaProperties,
					document: __props.document
				},
				schema: partitionedSchema.value.visibleProperties,
				schemaContext: "requestBody"
			}, null, 8, [
				"breadcrumb",
				"eventBus",
				"name",
				"options",
				"schema"
			]), createVNode(unref(Schema_default), {
				additionalProperties: "",
				breadcrumb: __props.breadcrumb,
				compact: "",
				compositionPath: ["requestBody"],
				eventBus: __props.eventBus,
				hideDescription: "",
				name: unref(translate)("operation.requestBody"),
				options: {
					hideReadOnly: true,
					orderRequiredPropertiesFirst: __props.options.orderRequiredPropertiesFirst,
					orderSchemaPropertiesBy: __props.options.orderSchemaPropertiesBy,
					expandAllSchemaProperties: __props.options.expandAllSchemaProperties,
					document: __props.document
				},
				schema: partitionedSchema.value.collapsedProperties,
				schemaContext: "requestBody"
			}, null, 8, [
				"breadcrumb",
				"eventBus",
				"name",
				"options",
				"schema"
			])])) : schema.value ? (openBlock(), createElementBlock("div", _hoisted_9, [createVNode(unref(Schema_default), {
				breadcrumb: __props.breadcrumb,
				compact: "",
				compositionPath: ["requestBody"],
				eventBus: __props.eventBus,
				hideReadOnly: true,
				name: unref(translate)("operation.requestBody"),
				noncollapsible: "",
				options: {
					hideReadOnly: true,
					orderRequiredPropertiesFirst: __props.options.orderRequiredPropertiesFirst,
					orderSchemaPropertiesBy: __props.options.orderSchemaPropertiesBy,
					expandAllSchemaProperties: __props.options.expandAllSchemaProperties,
					document: __props.document
				},
				schema: schema.value,
				schemaContext: "requestBody"
			}, null, 8, [
				"breadcrumb",
				"eventBus",
				"name",
				"options",
				"schema"
			])])) : createCommentVNode("", true)], 8, _hoisted_1)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { RequestBody_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=RequestBody.vue.script.js.map