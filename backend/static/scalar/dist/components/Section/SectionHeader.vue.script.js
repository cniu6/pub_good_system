import { createCommentVNode, createElementBlock, createElementVNode, defineComponent, normalizeClass, openBlock, renderSlot } from "vue";
//#region src/components/Section/SectionHeader.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "section-header-wrapper narrow:grid-cols-1 narrow:gap-0 grid grid-cols-2 gap-12" };
var SectionHeader_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SectionHeader",
	props: {
		tight: { type: Boolean },
		removeMargin: {
			type: Boolean,
			default: false
		}
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", _hoisted_1, [createElementVNode("div", { class: normalizeClass(["section-header", {
				tight: __props.tight,
				"mb-3": !__props.removeMargin
			}]) }, [renderSlot(_ctx.$slots, "default", {}, void 0, true)], 2), _ctx.$slots.links ? renderSlot(_ctx.$slots, "links", {}, void 0, true, 0) : createCommentVNode("", true)]);
		};
	}
});
//#endregion
export { SectionHeader_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SectionHeader.vue.script.js.map