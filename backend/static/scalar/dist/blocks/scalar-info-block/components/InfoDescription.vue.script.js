import InfoMarkdownSection_default from "./InfoMarkdownSection.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, defineComponent, openBlock, renderList } from "vue";
import { getHeadings, isHeading, splitContent, textFromNode } from "@scalar/code-highlight/markdown";
import { slugger } from "@scalar/helpers/string/slugger";
//#region src/blocks/scalar-info-block/components/InfoDescription.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "introduction-description mt-6 flex flex-col"
};
var InfoDescription_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "InfoDescription",
	props: {
		eventBus: {},
		headingSlugGenerator: { type: Function },
		description: {}
	},
	setup(__props) {
		/**
		* Descriptions, but split into multiple sections.
		* We need this to wrap the headings in IntersectionObserver components.
		*/
		const sections = computed(() => {
			if (!__props.description) return [];
			const { slug } = slugger();
			return splitContent(__props.description).map((markdown) => {
				const heading = getHeadings(markdown)[0];
				return {
					id: heading ? __props.headingSlugGenerator({
						...heading,
						slug: slug(heading.value)
					}) : void 0,
					content: markdown
				};
			});
		});
		/** Add ids to all headings */
		const transformHeading = (node) => {
			if (!isHeading(node)) return node;
			const { slug } = slugger();
			const value = textFromNode(node);
			node.data = { hProperties: { id: __props.headingSlugGenerator({
				depth: node.depth,
				value,
				slug: slug(value)
			}) } };
			return node;
		};
		return (_ctx, _cache) => {
			return __props.description ? (openBlock(), createElementBlock("div", _hoisted_1, [(openBlock(true), createElementBlock(Fragment, null, renderList(sections.value, (section) => {
				return openBlock(), createBlock(InfoMarkdownSection_default, {
					id: section.id,
					key: section.id,
					content: section.content,
					eventBus: __props.eventBus,
					transformHeading
				}, null, 8, [
					"id",
					"content",
					"eventBus"
				]);
			}), 128))])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { InfoDescription_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=InfoDescription.vue.script.js.map