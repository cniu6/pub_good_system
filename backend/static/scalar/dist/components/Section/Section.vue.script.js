import { useIntersection } from "../../hooks/use-intersection.js";
import { createElementBlock, defineComponent, openBlock, renderSlot, useTemplateRef } from "vue";
//#region src/components/Section/Section.vue?vue&type=script&setup=true&lang.ts
var Section_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Section",
	emits: ["intersecting"],
	setup(__props, { emit: __emit }) {
		const emit = __emit;
		const section = useTemplateRef("section");
		useIntersection(section, () => emit("intersecting"));
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("section", {
				ref_key: "section",
				ref: section,
				class: "section"
			}, [renderSlot(_ctx.$slots, "default", {}, void 0, true)], 512);
		};
	}
});
//#endregion
export { Section_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Section.vue.script.js.map