import { useIntersection } from "../../../hooks/use-intersection.js";
import { createElementBlock, createVNode, defineComponent, openBlock, unref, useTemplateRef } from "vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
//#region src/blocks/scalar-info-block/components/InfoMarkdownSection.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id"];
var InfoMarkdownSection_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "InfoMarkdownSection",
	props: {
		id: {},
		content: {},
		transformHeading: { type: Function },
		eventBus: {}
	},
	setup(__props) {
		const element = useTemplateRef("element");
		useIntersection(element, () => __props.id ? __props.eventBus?.emit("intersecting:nav-item", { id: __props.id }) : void 0);
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				id: __props.id,
				ref_key: "element",
				ref: element,
				class: "introduction-description-heading scroll-mt-16"
			}, [createVNode(unref(ScalarMarkdown), {
				transform: __props.transformHeading,
				transformType: "heading",
				value: __props.content,
				withImages: ""
			}, null, 8, ["transform", "value"])], 8, _hoisted_1);
		};
	}
});
//#endregion
export { InfoMarkdownSection_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=InfoMarkdownSection.vue.script.js.map