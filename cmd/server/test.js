const path = ["lolomo", 0, 0, ["title", "description"]]
const path2 = ["lolomo", 7, 7, ["title", "description"]]

datastore = {
    "lolomo": { 0: { type: "ref", value: ["list", "uuid"]}},
    "list": {
        "uuid": {
            0: { type: "ref", value: ["videos", 1234] },
        }
    },
    "videos": {
        1234: {
            "title": { type: "atom", value: "Live Free, Die Hard"},
            "description": "bruce willy givin it to us",
        }

    }
}

