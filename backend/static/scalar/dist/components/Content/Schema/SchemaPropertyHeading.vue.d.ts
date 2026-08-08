import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    value: SchemaObject | undefined;
    enum?: boolean;
    isDiscriminator?: boolean;
    required?: boolean;
    additional?: boolean;
    withExamples?: boolean;
    hideModelNames?: boolean;
    /** When the schema was resolved from a $ref, pass the ref name so it displays as e.g. "Data" instead of "object". */
    modelName?: string | null;
    /** Resolved propertyNames schema, used to surface key constraints like `format` for additional properties. */
    propertyNames?: SchemaObject;
    eventBus?: WorkspaceEventBus | null;
};
declare var __VLS_1: {};
type __VLS_Slots = {} & {
    name?: (props: typeof __VLS_1) => any;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {
    required: boolean;
    eventBus: WorkspaceEventBus | null;
    hideModelNames: boolean;
    isDiscriminator: boolean;
    withExamples: boolean;
}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=SchemaPropertyHeading.vue.d.ts.map