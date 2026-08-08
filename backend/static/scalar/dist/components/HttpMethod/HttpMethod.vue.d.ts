import type { HttpMethod } from '@scalar/helpers/http/http-methods';
import { type Component } from 'vue';
type __VLS_Props = {
    /** The type of element to render as, defaults to `span` */
    as?: Component | string;
    /** The css style property or variable that will be set to the request method color, defaults to `color` */
    property?: string;
    /** Whether or not to abbreviated the slot content */
    short?: boolean;
    /** The HTTP method to show */
    method: HttpMethod | string;
};
declare var __VLS_8: {};
type __VLS_Slots = {} & {
    default?: (props: typeof __VLS_8) => any;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=HttpMethod.vue.d.ts.map