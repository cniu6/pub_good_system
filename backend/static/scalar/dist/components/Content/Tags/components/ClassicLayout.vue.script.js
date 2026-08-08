import SectionContainerAccordion_default from "../../../Section/SectionContainerAccordion.vue.js";
import SectionHeader_default from "../../../Section/SectionHeader.vue.js";
import SectionHeaderTag_default from "../../../Section/SectionHeaderTag.vue.js";
import Anchor_default from "../../../Anchor/Anchor.vue.js";
import { createBlock, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, renderSlot, toDisplayString, unref, withCtx } from "vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/components/Content/Tags/components/ClassicLayout.vue?vue&type=script&setup=true&lang.ts
var ClassicLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ClassicLayout",
	props: {
		tag: {},
		isCollapsed: { type: Boolean },
		eventBus: {}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(SectionContainerAccordion_default), {
				"aria-label": __props.tag.title,
				class: normalizeClass(["tag-section", { "tag-section-group": __props.tag.isGroup }]),
				modelValue: !__props.isCollapsed,
				"onUpdate:modelValue": _cache[1] || (_cache[1] = (value) => __props.eventBus?.emit("toggle:nav-item", {
					id: __props.tag.id,
					open: value
				}))
			}, {
				title: withCtx(() => [createVNode(unref(SectionHeader_default), { class: normalizeClass(["tag-name", { "tag-group-name": __props.tag.isGroup }]) }, {
					default: withCtx(() => [createVNode(unref(Anchor_default), { onCopyAnchorUrl: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.tag.id })) }, {
						default: withCtx(() => [createVNode(unref(SectionHeaderTag_default), { level: 2 }, {
							default: withCtx(() => [createTextVNode(toDisplayString(__props.tag.title), 1)]),
							_: 1
						})]),
						_: 1
					})]),
					_: 1
				}, 8, ["class"]), createVNode(unref(ScalarMarkdown), {
					class: "tag-description",
					value: __props.tag?.description,
					withImages: ""
				}, null, 8, ["value"])]),
				default: withCtx(() => [renderSlot(_ctx.$slots, "default", {}, void 0, true)]),
				_: 3
			}, 8, [
				"aria-label",
				"class",
				"modelValue"
			]);
		};
	}
});
//#endregion
export { ClassicLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ClassicLayout.vue.script.js.map