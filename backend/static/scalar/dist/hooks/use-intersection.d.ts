import { type TemplateRef } from 'vue';
type UseIntersectionOptions = {
    /** Called when the element stops intersecting. */
    onExit?: () => void;
    /**
     * Whether to immediately fire the intersection as soon as the element is on the screen.
     *
     * When false (default) we use a root margin that is 2% of the viewport height on the middle of the viewport.
     */
    immediate?: boolean;
};
export declare const useIntersection: (el: TemplateRef<HTMLElement | undefined>, onIntersect: () => void, options?: UseIntersectionOptions) => void;
export {};
//# sourceMappingURL=use-intersection.d.ts.map