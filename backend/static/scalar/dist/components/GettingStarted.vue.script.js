import { useLocalization } from "../features/localization/use-localization.js";
import { Fragment, createElementBlock, createElementVNode, createStaticVNode, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, renderList, toDisplayString, unref, withCtx } from "vue";
import { themeLabels } from "@scalar/themes";
import { ScalarButton } from "@scalar/components/button";
//#region src/components/GettingStarted.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "custom-scroll start" };
var _hoisted_2 = { class: "start-copy" };
var _hoisted_3 = { class: "start-h1" };
var _hoisted_4 = { class: "start-p" };
var _hoisted_5 = { class: "start-cta" };
var _hoisted_6 = { class: "start-row" };
var _hoisted_7 = { class: "start-section start-section-integrations" };
var _hoisted_8 = { class: "start-h2" };
var _hoisted_9 = { class: "start-section start-section-colors" };
var _hoisted_10 = { class: "start-h2" };
var _hoisted_11 = ["onClick"];
var _hoisted_12 = { class: "start-h1" };
var _hoisted_13 = { class: "start-ul" };
var _hoisted_14 = { class: "start-h3" };
var _hoisted_15 = { class: "start-h3" };
var _hoisted_16 = { class: "start-h3" };
var _hoisted_17 = { class: "start-h3" };
var _hoisted_18 = { class: "start-h3" };
var _hoisted_19 = { class: "start-h3" };
var GettingStarted_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "GettingStarted",
	props: { theme: {} },
	emits: [
		"changeTheme",
		"loadSwaggerFile",
		"linkSwaggerFile",
		"updateContent"
	],
	setup(__props, { emit: __emit }) {
		const emits = __emit;
		const { translate } = useLocalization();
		const themeIds = [
			"default",
			"alternate",
			"moon",
			"purple",
			"solarized",
			"bluePlanet",
			"saturn",
			"kepler",
			"mars",
			"deepSpace"
		];
		async function fetchExampleSpecification() {
			const response = await fetch("https://registry.scalar.com/@scalar/apis/galaxy?format=yaml");
			emits("updateContent", await response.text());
		}
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", _hoisted_1, [
				createElementVNode("div", _hoisted_2, [
					_cache[1] || (_cache[1] = createElementVNode("div", { class: "start-logo" }, [createElementVNode("svg", {
						height: "36",
						viewBox: "0 0 36 36",
						width: "36",
						xmlns: "http://www.w3.org/2000/svg"
					}, [createElementVNode("path", {
						d: "M18 0a18 18 0 1 1 0 36 18 18 0 0 1 0-36Zm11.2 6.5c-3.3-3.3-11.1-1-17.4 5.3-6.2 6.3-8.6 14-5.3 17.4 3.3 3.3 11 .9 17.3-5.4 6.3-6.2 8.7-14 5.4-17.3ZM17.6 12a6.3 6.3 0 1 1 0 12.7 6.3 6.3 0 0 1 0-12.7Z",
						fill: "currentColor",
						"fill-rule": "evenodd"
					})])], -1)),
					createElementVNode("h1", _hoisted_3, toDisplayString(unref(translate)("gettingStarted.swaggerEditor")), 1),
					createElementVNode("p", _hoisted_4, toDisplayString(unref(translate)("gettingStarted.description")), 1),
					createElementVNode("div", _hoisted_5, [createVNode(unref(ScalarButton), {
						class: "w-full",
						onClick: fetchExampleSpecification
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("gettingStarted.showExample")), 1)]),
						_: 1
					}), createVNode(unref(ScalarButton), {
						class: "w-full",
						variant: "outlined",
						onClick: _cache[0] || (_cache[0] = ($event) => _ctx.$emit("loadSwaggerFile"))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("gettingStarted.uploadFile")), 1)]),
						_: 1
					})])
				]),
				createElementVNode("div", _hoisted_6, [createElementVNode("div", _hoisted_7, [createElementVNode("div", _hoisted_8, toDisplayString(unref(translate)("gettingStarted.integrations")), 1), _cache[2] || (_cache[2] = createStaticVNode("<a class=\"start-item\" href=\"https://github.com/scalar/scalar/tree/main/integrations/fastify#readme\" rel=\"noopener noreferrer\" target=\"_blank\" data-v-c492e215><svg fill=\"currentColor\" height=\"16\" viewBox=\"0 0 19 16\" width=\"19\" xmlns=\"http://www.w3.org/2000/svg\" data-v-c492e215><path d=\"m18.2 3.1.8-2V.9l-4.7 1.3C15.2 1 15 0 15 0s-2.5 1.6-4.3 1.5c-2 0-3.6.8-4 1-1.8 1.2-2.5 3.3-3.2 3.8L0 8.9 2.3 8l-2 2.5c.2.3 1.2 1.6 2.1 1.3l.4-.1 1.6.5-.7-1 .2-.2.9.3-.1-.8.9.3-.1-.8.3-.1 1-3.5 3.7-2.6-.3.7A4 4 0 0 1 8 7l-.6.2c-.5.5-.7.7-.8 2.5a2 2 0 0 1 1 0c1.6.4 2.2 2.3 1.7 2.9l-.7.6H8v.6h-.7v.5l-.2.2c-.7 0-1.4-.6-1.4-.6 0 .5.4 1.3.4 1.3s1.7 1.1 2.7.7c1-.4.7-2.3 2.8-3.2l3.3-.9.8-2.2-1.7.5v-2l2.5-.6.9-2.2-3.4.9v-2l4.2-1.1Z\" fill=\"currentColor\" fill-rule=\"nonzero\" data-v-c492e215></path></svg><span data-v-c492e215>Fastify</span></a><a class=\"start-item\" href=\"https://github.com/scalar/scalar/blob/main/documentation/integrations/html-js.md#html\" rel=\"noopener noreferrer\" target=\"_blank\" data-v-c492e215><svg fill=\"currentColor\" viewBox=\"0 0 24 24\" xmlns=\"http://www.w3.org/2000/svg\" data-v-c492e215><g data-v-c492e215><path d=\"M22.5 1.5h-21A1.5 1.5 0 0 0 0 3v3a1.5 1.5 0 0 0 1.5 1.5h21A1.5 1.5 0 0 0 24 6V3a1.5 1.5 0 0 0-1.5-1.5Zm-19.25 3A1.25 1.25 0 1 1 4.5 5.75 1.25 1.25 0 0 1 3.25 4.5ZM8.5 5.75A1.25 1.25 0 1 1 9.75 4.5 1.25 1.25 0 0 1 8.5 5.75Z\" fill=\"currentColor\" data-v-c492e215></path><path d=\"M22.5 9h-21A1.5 1.5 0 0 0 0 10.5v3A1.5 1.5 0 0 0 1.5 15h21a1.5 1.5 0 0 0 1.5-1.5v-3A1.5 1.5 0 0 0 22.5 9ZM3.25 12a1.25 1.25 0 1 1 1.25 1.25A1.25 1.25 0 0 1 3.25 12Zm5.25 1.25A1.25 1.25 0 1 1 9.75 12a1.25 1.25 0 0 1-1.25 1.25Z\" fill=\"currentColor\" data-v-c492e215></path><path d=\"M22.5 16.5h-21A1.5 1.5 0 0 0 0 18v3a1.5 1.5 0 0 0 1.5 1.5h21A1.5 1.5 0 0 0 24 21v-3a1.5 1.5 0 0 0-1.5-1.5Zm-19.25 3a1.25 1.25 0 1 1 1.25 1.25 1.25 1.25 0 0 1-1.25-1.25Zm5.25 1.25a1.25 1.25 0 1 1 1.25-1.25 1.25 1.25 0 0 1-1.25 1.25Z\" fill=\"currentColor\" data-v-c492e215></path></g></svg><span data-v-c492e215>CDN</span></a><a class=\"start-item\" href=\"https://github.com/scalar/scalar/blob/main/packages/api-reference/README.md#vuejs\" rel=\"noopener noreferrer\" target=\"_blank\" data-v-c492e215><svg height=\"170\" viewBox=\"0 0 196.3 170\" width=\"196.3\" xmlns=\"http://www.w3.org/2000/svg\" data-v-c492e215><g fill=\"currentColor\" fill-rule=\"nonzero\" data-v-c492e215><polygon points=\"39.23 0 0 0 2.9450761 5.1010782 98.16 170.02 196.32 0 157.06 0 98.16 102.01 42.175701 5.0991171\" data-v-c492e215></polygon><polygon points=\"75.5 2.009956e-14 0 2.009956e-14 2.94 5.1 78.44871 5.1 98.16 39.26 117.87937 5.1 193.38 5.1 196.325 0 120.82 7.8065636e-15 114.97322 2.009956e-14 98.16 29.037153 81.35 2.009956e-14\" data-v-c492e215></polygon></g></svg><span data-v-c492e215>Vue</span></a><a class=\"start-item\" href=\"https://github.com/scalar/scalar/blob/main/packages/api-reference-react/README.md#usage\" rel=\"noopener noreferrer\" target=\"_blank\" data-v-c492e215><svg height=\"23.3\" viewBox=\"0 0 22 23.3\" width=\"22\" xmlns=\"http://www.w3.org/2000/svg\" data-v-c492e215><g fill=\"none\" fill-rule=\"evenodd\" data-v-c492e215><circle cx=\"11\" cy=\"11.6\" fill=\"currentColor\" fill-rule=\"nonzero\" r=\"2\" data-v-c492e215></circle><g stroke=\"currentColor\" data-v-c492e215><ellipse cx=\"11\" cy=\"11.6\" rx=\"11\" ry=\"4.2\" data-v-c492e215></ellipse><ellipse cx=\"11\" cy=\"11.6\" rx=\"11\" ry=\"4.2\" transform=\"rotate(60 11 11.6)\" data-v-c492e215></ellipse><ellipse cx=\"11\" cy=\"11.6\" rx=\"11\" ry=\"4.2\" transform=\"rotate(120 11 11.6)\" data-v-c492e215></ellipse></g></g></svg><span data-v-c492e215>React</span></a>", 4))]), createElementVNode("div", _hoisted_9, [createElementVNode("p", _hoisted_10, toDisplayString(unref(translate)("gettingStarted.theming")), 1), (openBlock(), createElementBlock(Fragment, null, renderList(themeIds, (themeId) => {
					return createElementVNode("div", {
						key: themeId,
						class: normalizeClass(["start-item", { "start-item-active": themeId === __props.theme }]),
						onClick: ($event) => _ctx.$emit("changeTheme", {
							id: themeId,
							label: unref(themeLabels)[themeId]
						})
					}, toDisplayString(unref(themeLabels)[themeId]), 11, _hoisted_11);
				}), 64))])]),
				createElementVNode("p", _hoisted_12, toDisplayString(unref(translate)("gettingStarted.features")), 1),
				createElementVNode("ul", _hoisted_13, [
					createElementVNode("li", null, [createElementVNode("p", _hoisted_14, toDisplayString(unref(translate)("gettingStarted.customize")), 1), createTextVNode(" " + toDisplayString(unref(translate)("gettingStarted.customizeDescription")), 1)]),
					createElementVNode("li", null, [createElementVNode("p", _hoisted_15, toDisplayString(unref(translate)("gettingStarted.testing")), 1), createTextVNode(" " + toDisplayString(unref(translate)("gettingStarted.testingDescription")), 1)]),
					createElementVNode("li", null, [createElementVNode("p", _hoisted_16, toDisplayString(unref(translate)("gettingStarted.search")), 1), createTextVNode(" " + toDisplayString(unref(translate)("gettingStarted.searchDescription")), 1)]),
					createElementVNode("li", null, [createElementVNode("p", _hoisted_17, toDisplayString(unref(translate)("gettingStarted.hosting")), 1), createTextVNode(" " + toDisplayString(unref(translate)("gettingStarted.hostingDescription")), 1)]),
					createElementVNode("li", null, [createElementVNode("p", _hoisted_18, toDisplayString(unref(translate)("gettingStarted.openApiSwagger")), 1), createTextVNode(" " + toDisplayString(unref(translate)("gettingStarted.openApiSwaggerDescription")), 1)]),
					createElementVNode("li", null, [createElementVNode("p", _hoisted_19, toDisplayString(unref(translate)("gettingStarted.codeSamples")), 1), createTextVNode(" " + toDisplayString(unref(translate)("gettingStarted.codeSamplesDescription")), 1)])
				])
			]);
		};
	}
});
//#endregion
export { GettingStarted_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=GettingStarted.vue.script.js.map