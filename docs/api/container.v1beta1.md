<h1>Source API reference</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#container.kjournal%2fv1beta1">container.kjournal/v1beta1</a>
</li>
</ul>
<h2 id="container.kjournal/v1beta1">container.kjournal/v1beta1</h2>
Resource Types:
<ul class="simple"><li>
<a href="#container.kjournal/v1beta1.Log">Log</a>
</li></ul>
<h3 id="container.kjournal/v1beta1.Log">Log
</h3>
<p>Log</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br>
string</td>
<td>
<code>container.kjournal/v1beta1</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
string
</td>
<td>
<code>Log</code>
</td>
</tr>
<tr>
<td>
<code>-</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
<p>ObjectMeta is only included to fullfil metav1.Object interface,
it will be omitted from any json de and encoding. It is required for storage.ConvertToTable()</p>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="#container.kjournal/v1beta1.LogMetadata">
LogMetadata
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>container</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pod</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>unstructured</code><br>
<em>
encoding/json.RawMessage
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>env</code><br>
<em>
encoding/json.RawMessage
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="container.kjournal/v1beta1.LogMetadata">LogMetadata
</h3>
<p>
(<em>Appears on:</em>
<a href="#container.kjournal/v1beta1.Log">Log</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>namespace</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>creationTimestamp</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
