var id = "http://rdf.freebase.com/ns/m.035dk";
var rel = G.Bs("http://rdf.freebase.com/ns/type.object.name")
var i = G.AllOut().Out(rel).Walk(G.Graph(), G.Vertex(id)).Iter(10);
var acc = []
for (var x = i.Next(); !i.IsClosed(); x = i.Next()) {
    if (!x[0]) { break; }
    var tuple = [x[0].ToStrings()[1], x[1].Strings()[2]];
    acc.push(tuple);
}
acc;
