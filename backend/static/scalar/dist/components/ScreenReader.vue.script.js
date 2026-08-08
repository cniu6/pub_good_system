import { createElementBlock, defineComponent, openBlock, renderSlot } from "vue";
//#region src/components/ScreenReader.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "screenreader-only"
};
var ScreenReader_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ScreenReader",
	props: { if: {
		type: Boolean,
		default: true
	} },
	setup(__props) {
		return (_ctx, _cache) => {
			return _ctx.$props.if ? (openBlock(), createElementBlock("span", _hoisted_1, [renderSlot(_ctx.$slots, "default", {}, void 0, true)])) : renderSlot(_ctx.$slots, "default", {}, void 0, true, 1);
		};
	}
});
//#endregion
export { ScreenReader_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ScreenReader.vue.script.js.map