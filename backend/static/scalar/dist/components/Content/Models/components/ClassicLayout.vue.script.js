import SectionAccordion_default from "../../../Section/SectionAccordion.vue.js";
import SectionHeaderTag_default from "../../../Section/SectionHeaderTag.vue.js";
import Anchor_default from "../../../Anchor/Anchor.vue.js";
import SchemaHeading_default from "../../Schema/SchemaHeading.vue.js";
import SchemaProperty_default from "../../Schema/SchemaProperty.vue.js";
import { Fragment, createBlock, createElementBlock, createVNode, defineComponent, openBlock, renderList, unref, withCtx } from "vue";
import { resolve } from "@scalar/workspace-store/resolve";
//#region src/components/Content/Models/components/ClassicLayout.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "properties"
};
var _hoisted_2 = { key: 1 };
var ClassicLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ClassicLayout",
	props: {
		id: {},
		name: {},
		schema: {},
		isCollapsed: { type: Boolean },
		eventBus: {},
		document: {},
		options: {}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(SectionAccordion_default), {
				"aria-label": __props.schema.title ?? __props.name,
				modelValue: !__props.isCollapsed,
				"onUpdate:modelValue": _cache[1] || (_cache[1] = (value) => __props.eventBus?.emit("toggle:nav-item", {
					id: __props.id,
					open: value
				}))
			}, {
				title: withCtx(() => [createVNode(unref(Anchor_default), {
					class: "reference-models-anchor",
					eventBus: __props.eventBus,
					onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.id }))
				}, {
					default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), { level: 3 }, {
						default: withCtx(() => [createVNode(unref(SchemaHeading_default), {
							class: "reference-models-label",
							name: __props.schema.title ?? __props.name,
							value: __props.schema
						}, null, 8, ["name", "value"])]),
						_: 1
					})]),
					_: 1
				}, 8, ["eventBus"])]),
				default: withCtx(() => ["properties" in __props.schema ? (openBlock(), createElementBlock("div", _hoisted_1, [(openBlock(true), createElementBlock(Fragment, null, renderList(Object.entries(__props.schema.properties ?? {}), ([property, value]) => {
					return openBlock(), createBlock(unref(SchemaProperty_default), {
						key: property,
						eventBus: __props.eventBus,
						hideModelNames: __props.options.hideModels,
						name: property,
						options: {
							...__props.options,
							document: __props.document
						},
						required: __props.schema.required?.includes(property),
						schema: unref(resolve).schema(value)
					}, null, 8, [
						"eventBus",
						"hideModelNames",
						"name",
						"options",
						"required",
						"schema"
					]);
				}), 128))])) : (openBlock(), createElementBlock("div", _hoisted_2, [createVNode(unref(SchemaProperty_default), {
					eventBus: __props.eventBus,
					hideModelNames: __props.options.hideModels,
					options: {
						...__props.options,
						document: __props.document
					},
					schema: __props.schema
				}, null, 8, [
					"eventBus",
					"hideModelNames",
					"options",
					"schema"
				])]))]),
				_: 1
			}, 8, ["aria-label", "modelValue"]);
		};
	}
});
//#endregion
export { ClassicLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ClassicLayout.vue.script.js.map