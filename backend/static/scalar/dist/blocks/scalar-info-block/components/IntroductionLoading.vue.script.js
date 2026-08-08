import SectionColumn_default from "../../../components/Section/SectionColumn.vue.js";
import SectionColumns_default from "../../../components/Section/SectionColumns.vue.js";
import { createBlock, createCommentVNode, createElementBlock, createElementVNode, createStaticVNode, createVNode, defineComponent, openBlock, unref, withCtx } from "vue";
//#region src/blocks/scalar-info-block/components/IntroductionLoading.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	"aria-hidden": "true",
	class: "introduction-loading flex flex-col gap-5"
};
var _hoisted_2 = {
	key: 0,
	class: "narrow:flex-col narrow:gap-3 flex gap-6"
};
var IntroductionLoading_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "IntroductionLoading",
	props: { hasAside: {
		type: Boolean,
		default: true
	} },
	setup(__props) {
		/**
		* Loading skeleton for the introduction block. It mirrors the real layout
		* (badges, title, links, description and selector cards) so the page does not
		* jump once the document has loaded.
		*/
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", _hoisted_1, [
				_cache[3] || (_cache[3] = createStaticVNode("<div class=\"flex gap-1.5\" data-v-41a61989><div class=\"introduction-skeleton h-6 w-14 rounded-full\" data-v-41a61989></div><div class=\"introduction-skeleton h-6 w-24 rounded-full\" data-v-41a61989></div></div><div class=\"narrow:grid-cols-1 narrow:gap-3 grid grid-cols-2 gap-12\" data-v-41a61989><div class=\"introduction-skeleton h-9 w-3/5 rounded-lg\" data-v-41a61989></div><div class=\"narrow:justify-start flex flex-wrap items-center justify-end gap-2\" data-v-41a61989><div class=\"introduction-skeleton h-5 w-28 rounded\" data-v-41a61989></div><div class=\"introduction-skeleton h-5 w-28 rounded\" data-v-41a61989></div><div class=\"introduction-skeleton h-5 w-12 rounded\" data-v-41a61989></div></div></div>", 2)),
				createVNode(unref(SectionColumns_default), null, {
					default: withCtx(() => [createVNode(unref(SectionColumn_default), null, {
						default: withCtx(() => [..._cache[0] || (_cache[0] = [createElementVNode("div", { class: "flex flex-col gap-3" }, [
							createElementVNode("div", { class: "introduction-skeleton mb-2 h-5 w-56 rounded" }),
							createElementVNode("div", { class: "introduction-skeleton h-4 w-full rounded" }),
							createElementVNode("div", { class: "introduction-skeleton h-4 w-11/12 rounded" }),
							createElementVNode("div", { class: "introduction-skeleton h-4 w-4/5 rounded" }),
							createElementVNode("div", { class: "introduction-skeleton mt-4 h-6 w-40 rounded" }),
							createElementVNode("div", { class: "introduction-skeleton h-4 w-3/4 rounded" }),
							createElementVNode("div", { class: "introduction-skeleton h-4 w-2/3 rounded" }),
							createElementVNode("div", { class: "introduction-skeleton h-4 w-1/2 rounded" })
						], -1)])]),
						_: 1
					}), __props.hasAside ? (openBlock(), createBlock(unref(SectionColumn_default), { key: 0 }, {
						default: withCtx(() => [..._cache[1] || (_cache[1] = [createElementVNode("div", { class: "sticky-cards gap-3" }, [
							createElementVNode("div", { class: "introduction-skeleton h-20 w-full rounded-lg" }),
							createElementVNode("div", { class: "introduction-skeleton h-28 w-full rounded-lg" }),
							createElementVNode("div", { class: "introduction-skeleton h-28 w-full rounded-lg" })
						], -1)])]),
						_: 1
					})) : createCommentVNode("", true)]),
					_: 1
				}),
				!__props.hasAside ? (openBlock(), createElementBlock("div", _hoisted_2, [..._cache[2] || (_cache[2] = [
					createElementVNode("div", { class: "introduction-skeleton h-28 w-full flex-1 rounded-lg" }, null, -1),
					createElementVNode("div", { class: "introduction-skeleton h-28 w-full flex-1 rounded-lg" }, null, -1),
					createElementVNode("div", { class: "introduction-skeleton h-28 w-full flex-1 rounded-lg" }, null, -1)
				])])) : createCommentVNode("", true)
			]);
		};
	}
});
//#endregion
export { IntroductionLoading_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=IntroductionLoading.vue.script.js.map