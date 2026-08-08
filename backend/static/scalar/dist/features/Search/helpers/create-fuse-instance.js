import Fuse from "fuse.js";
//#region src/features/Search/helpers/create-fuse-instance.ts
/**
* Create a Fuse instance for searching the API reference.
*
* Doesn't have any data yet, so it's empty.
*/
function createFuseInstance() {
	return new Fuse([], {
		keys: [
			{
				name: "title",
				weight: .7
			},
			{
				name: "operationId",
				weight: .6
			},
			{
				name: "parameters",
				weight: .55
			},
			{
				name: "body",
				weight: .55
			},
			{
				name: "path",
				weight: .5
			},
			{
				name: "tag",
				weight: .4
			},
			{
				name: "description",
				weight: .3
			},
			{
				name: "method",
				weight: .3
			},
			{
				name: "responseExamples",
				weight: .25
			},
			{
				name: "parameterDescriptions",
				weight: .2
			},
			{
				name: "bodyDescriptions",
				weight: .2
			}
		],
		threshold: .3,
		distance: 100,
		includeScore: true,
		includeMatches: true,
		ignoreLocation: true,
		useExtendedSearch: true,
		findAllMatches: true
	});
}
//#endregion
export { createFuseInstance };

//# sourceMappingURL=create-fuse-instance.js.map