import SectionHeaderTag_default from "../../../Section/SectionHeaderTag.vue.js";
import CompactSection_default from "../../../Section/CompactSection.vue.js";
import SchemaHeading_default from "../../Schema/SchemaHeading.vue.js";
import Schema_default from "../../Schema/Schema.vue.js";
import { createBlock, createVNode, defineComponent, openBlock, unref, withCtx } from "vue";
import { ScalarErrorBoundary } from "@scalar/components/error-boundary";
//#region src/components/Content/Models/components/ModernLayout.vue?vue&type=script&setup=true&lang.ts
var ModernLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ModernLayout",
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
			return openBlock(), createBlock(unref(CompactSection_default), {
				id: __props.id,
				key: __props.name,
				label: __props.name,
				modelValue: !__props.isCollapsed,
				onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.id })),
				"onUpdate:modelValue": _cache[1] || (_cache[1] = (value) => __props.eventBus?.emit("toggle:nav-item", {
					id: __props.id,
					open: value
				}))
			}, {
				heading: withCtx(() => [createVNode(unref(SectionHeaderTag_default), { level: 3 }, {
					default: withCtx(() => [createVNode(unref(SchemaHeading_default), {
						name: __props.schema.title ?? __props.name,
						value: __props.schema
					}, null, 8, ["name", "value"])]),
					_: 1
				})]),
				default: withCtx(() => [createVNode(unref(ScalarErrorBoundary), null, {
					default: withCtx(() => [createVNode(unref(Schema_default), {
						eventBus: __props.eventBus,
						hideModelNames: __props.options.hideModels,
						hideHeading: "",
						level: 1,
						noncollapsible: "",
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
					])]),
					_: 1
				})]),
				_: 1
			}, 8, [
				"id",
				"label",
				"modelValue"
			]);
		};
	}
});
//#endregion
export { ModernLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ModernLayout.vue.script.js.map