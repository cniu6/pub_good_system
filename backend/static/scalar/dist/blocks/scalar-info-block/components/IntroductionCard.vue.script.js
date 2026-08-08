import { createElementBlock, defineComponent, normalizeClass, openBlock, renderSlot } from "vue";
//#region src/blocks/scalar-info-block/components/IntroductionCard.vue?vue&type=script&setup=true&lang.ts
var IntroductionCard_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "IntroductionCard",
	props: { row: { type: Boolean } },
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", { class: normalizeClass(["introduction-card", { "introduction-card-row": __props.row }]) }, [renderSlot(_ctx.$slots, "default", {}, void 0, true)], 2);
		};
	}
});
//#endregion
export { IntroductionCard_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=IntroductionCard.vue.script.js.map