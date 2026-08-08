import { createBlock, defineComponent, openBlock, renderSlot, resolveDynamicComponent, withCtx } from "vue";
//#region src/components/Section/SectionHeaderTag.vue?vue&type=script&setup=true&lang.ts
var SectionHeaderTag_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SectionHeaderTag",
	props: { level: { default: 1 } },
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createBlock(resolveDynamicComponent(`h${__props.level}`), { class: "section-header-label" }, {
				default: withCtx(() => [renderSlot(_ctx.$slots, "default", {}, void 0, true)]),
				_: 3
			});
		};
	}
});
//#endregion
export { SectionHeaderTag_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SectionHeaderTag.vue.script.js.map