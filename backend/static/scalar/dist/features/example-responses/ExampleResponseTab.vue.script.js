import { createBlock, createElementVNode, defineComponent, normalizeClass, openBlock, renderSlot, unref, withCtx } from "vue";
import { Tab } from "@headlessui/vue";
//#region src/features/example-responses/ExampleResponseTab.vue?vue&type=script&setup=true&lang.ts
var ExampleResponseTab_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ExampleResponseTab",
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(Tab), { as: "template" }, {
				default: withCtx(({ selected }) => [createElementVNode("button", {
					class: normalizeClass(["tab", { "tab-selected": selected }]),
					type: "button"
				}, [createElementVNode("span", null, [renderSlot(_ctx.$slots, "default", {}, void 0, true)])], 2)]),
				_: 3
			});
		};
	}
});
//#endregion
export { ExampleResponseTab_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ExampleResponseTab.vue.script.js.map