function holonyms(term) {
    var label = G.Bs("http://www.w3.org/2000/01/rdf-schema#label");
    var holo = G.Bs("http://wordnet-rdf.princeton.edu/ontology#part_holonym");
    var paths = G.In(label).Out(holo).Out(label).Walk(G.Graph(), G.Vertex(term)).Collect();
    var uniq = {};
    var acc = [];
    for (var i=0; i<paths.length; i++) {
	var h = paths[i][2].Strings()[2];
	console.log(h);
	if (!uniq[h]) {
            uniq[h] = true;
	    acc.push(h);
	}
    }
    return acc;
}
holonyms("Africa");
