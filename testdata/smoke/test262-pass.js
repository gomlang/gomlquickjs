/*---
flags: [raw]
---*/
if (JSON.stringify([1, 2, 3].map(x => x * 2)) !== '[2,4,6]') {
    throw new Error('array mapping regression');
}
