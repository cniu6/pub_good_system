import { mergeAllOfSchemas } from "./merge-all-of-schemas.js";
import { resolve } from "@scalar/workspace-store/resolve";
//#region src/components/Content/Schema/helpers/partition-all-of-compositions.ts
var CHOICE_KEYWORDS = ["oneOf", "anyOf"];
/**
* Flattens a member's own (non-composition) properties and every `oneOf`/`anyOf`
* group it declares — recursing into nested `allOf` — into a flat, ordered list,
* preserving source order so choice groups stay where they were authored.
*
* Pure-constraint keywords (`not`, `if/then/else`) are dropped: they carry no
* visual variant and Scalar would otherwise render `not` as a bogus picker. The
* rule still lives in the schema (validation) and in field descriptions.
*/
var collectMembers = (schema, out, seenRefs) => {
	const { allOf, oneOf, anyOf, not: _not, if: _if, then: _then, else: _else, ...rest } = schema;
	if (Object.keys(rest).length > 0) out.push({
		kind: "object",
		schema: rest
	});
	for (const keyword of CHOICE_KEYWORDS) {
		const value = keyword === "oneOf" ? oneOf : anyOf;
		if (Array.isArray(value) && value.length > 0) out.push({
			kind: "choice",
			composition: keyword,
			value: { [keyword]: value }
		});
	}
	if (Array.isArray(allOf)) {
		for (const rawMember of allOf) if (rawMember && typeof rawMember === "object") {
			const resolved = resolve.schema(rawMember);
			const ref = resolved.$ref;
			if (typeof ref === "string") {
				if (seenRefs.has(ref)) continue;
				collectMembers(resolved, out, new Set(seenRefs).add(ref));
			} else collectMembers(resolved, out, seenRefs);
		}
	}
};
/**
* Splits an `allOf` schema into an ordered list of segments so the renderer can
* show each `oneOf`/`anyOf` group as its own picker **in the position it was
* declared**, with the surrounding plain fields around it.
*
* `mergeAllOfSchemas` alone keeps only the FIRST `oneOf`/`anyOf` (dropping the
* 2nd+ groups of an object with several independent mutually-exclusive
* selections) and, being a merge, also loses the ordering between fields and
* choices. Walking the members in order fixes both: runs of consecutive object
* members are merged into one object segment, and each choice member becomes its
* own picker segment in place.
*/
var partitionAllOfCompositions = (schema) => {
	if (!schema) return { segments: [] };
	const { allOf, oneOf: _oneOf, anyOf: _anyOf, not: _not, if: _if, then: _then, else: _else, ...rest } = schema;
	if (!Array.isArray(allOf)) return { segments: [{
		kind: "object",
		schema
	}] };
	const members = [];
	if (Object.keys(rest).length > 0) members.push({
		kind: "object",
		schema: rest
	});
	const seenRefs = /* @__PURE__ */ new Set();
	for (const rawMember of allOf) if (rawMember && typeof rawMember === "object") {
		const resolved = resolve.schema(rawMember);
		const ref = resolved.$ref;
		collectMembers(resolved, members, typeof ref === "string" ? new Set(seenRefs).add(ref) : seenRefs);
	}
	const segments = [];
	let objectRun = [];
	let choiceIndex = 0;
	const flushObjectRun = () => {
		if (objectRun.length === 0) return;
		const merged = objectRun.length === 1 ? objectRun[0] : mergeAllOfSchemas({ allOf: objectRun });
		segments.push({
			kind: "object",
			schema: merged
		});
		objectRun = [];
	};
	for (const member of members) if (member.kind === "object") objectRun.push(member.schema);
	else {
		flushObjectRun();
		segments.push({
			kind: "choice",
			composition: member.composition,
			value: member.value,
			choiceIndex: choiceIndex++
		});
	}
	flushObjectRun();
	return { segments };
};
//#endregion
export { partitionAllOfCompositions };

//# sourceMappingURL=partition-all-of-compositions.js.map