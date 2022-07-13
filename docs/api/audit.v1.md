<h1>Source API reference</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#audit.kjournal%2fv1">audit.kjournal/v1</a>
</li>
</ul>
<h2 id="audit.kjournal/v1">audit.kjournal/v1</h2>
Resource Types:
<ul class="simple"><li>
<a href="#audit.kjournal/v1.ClusterEvent">ClusterEvent</a>
</li></ul>
<h3 id="audit.kjournal/v1.ClusterEvent">ClusterEvent
</h3>
<p>ClusterEvent</p>
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
<code>audit.kjournal/v1</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
string
</td>
<td>
<code>ClusterEvent</code>
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
<code>Event</code><br>
<em>
k8s.io/apiserver/pkg/apis/audit/v1.Event
</em>
</td>
<td>
<p>
(Members of <code>Event</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="audit.kjournal/v1.Event">Event
</h3>
<p>Event</p>
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
<code>Event</code><br>
<em>
k8s.io/apiserver/pkg/apis/audit/v1.Event
</em>
</td>
<td>
<p>
(Members of <code>Event</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
