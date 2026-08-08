import { createBlock, createVNode, defineComponent, openBlock, renderSlot, unref, withCtx } from "vue";
import { TabGroup, TabList } from "@headlessui/vue";
import { ScalarCardHeader } from "@scalar/components/card";
//#region src/features/example-responses/ExampleResponseTabList.vue?vue&type=script&setup=true&lang.ts
var ExampleResponseTabList_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ExampleResponseTabList",
	emits: ["change"],
	setup(__props, { emit: __emit }) {
		const emit = __emit;
		const changeTab = (index) => {
			emit("change", index);
		};
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarCardHeader), { class: "scalar-card-header scalar-card-header-tabs" }, {
				actions: withCtx(() => [renderSlot(_ctx.$slots, "actions", {}, void 0, true)]),
				default: withCtx(() => [createVNode(unref(TabGroup), { onChange: changeTab }, {
					default: withCtx(() => [createVNode(unref(TabList), { class: "tab-list custom-scroll" }, {
						default: withCtx(() => [renderSlot(_ctx.$slots, "default", {}, void 0, true)]),
						_: 3
					})]),
					_: 3
				})]),
				_: 3
			});
		};
	}
});
//#endregion
export { ExampleResponseTabList_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ExampleResponseTabList.vue.script.js.map