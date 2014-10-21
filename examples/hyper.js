var label = G.Bs("http://www.w3.org/2000/01/rdf-schema#label");

function find(term) {
    var paths = G.In(label).Walk(G.Graph(), G.Vertex(term)).Collect();
    var acc = [];
    for (var i=0; i<paths.length; i++) {
	var id = paths[i][0].Strings()[2];
	acc.push(id);
    }
    return acc;
}

function labels(id) {
    var paths = G.Out(label).Walk(G.Graph(), G.Vertex(id)).Collect();
    var acc = [];
    for (var i=0; i<paths.length; i++) {
	var name = paths[i][0].Strings()[2];
	acc.push(name);
    }
    return acc;
}

function collect(rel, id, acc, uniq, recursive, reverse, maxDepth, depth) {
    console.log("collect", id, "depth", depth, "labels", labels(id));
    var paths;
    if (reverse) {
	paths = G.In(rel).Walk(G.Graph(), G.Vertex(id)).Collect();
    } else {
	paths = G.Out(rel).Walk(G.Graph(), G.Vertex(id)).Collect();
    }
    for (var i=0; i<paths.length; i++) {
	var h = paths[i][0].Strings()[2];
	if (!uniq[h]) {
            uniq[h] = true;
	    acc.push({labels: labels(h), depth: depth});
	    if (recursive && depth <= maxDepth) {
		collect(rel, h, acc, uniq, recursive, reverse, maxDepth, depth+1);
	    }
	}
    }
}

function recurse(rel, reverse, maxDepth, term) {
    var acc = [];
    var uniq = {};
    var ids = find(term);
    for (var i=0; i<ids.length; i++) {
	collect(G.Bs(rel), ids[i], acc, uniq, true, reverse, maxDepth, 0);
    }
    return acc;
}

function hypernyms(term) {
    return recurse("http://wordnet-rdf.princeton.edu/ontology#hypernym", false, 1000, term);
}

hypernyms("virus");
