import { createElementBlock, defineComponent, openBlock, renderSlot } from "vue";
//#region src/components/Section/SectionContainer.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 1,
	class: "section-container"
};
var SectionContainer_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SectionContainer",
	props: { omit: { type: Boolean } },
	setup(__props) {
		return (_ctx, _cache) => {
			return __props.omit ? renderSlot(_ctx.$slots, "default", {}, void 0, true, 0) : (openBlock(), createElementBlock("div", _hoisted_1, [renderSlot(_ctx.$slots, "default", {}, void 0, true)]));
		};
	}
});
//#endregion
export { SectionContainer_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SectionContainer.vue.script.js.map