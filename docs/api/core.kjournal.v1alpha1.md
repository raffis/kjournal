<h1>Source API reference</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#core.kjournal%2fv1alpha1">core.kjournal/v1alpha1</a>
</li>
</ul>
<h2 id="core.kjournal/v1alpha1">core.kjournal/v1alpha1</h2>
Resource Types:
<ul class="simple"><li>
<a href="#core.kjournal/v1alpha1.AuditEvent">AuditEvent</a>
</li><li>
<a href="#core.kjournal/v1alpha1.Event">Event</a>
</li><li>
<a href="#core.kjournal/v1alpha1.Log">Log</a>
</li></ul>
<h3 id="core.kjournal/v1alpha1.AuditEvent">AuditEvent
</h3>
<p>AuditEvent</p>
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
<code>core.kjournal/v1alpha1</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
string
</td>
<td>
<code>AuditEvent</code>
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
k8s.io/apiserver/pkg/apis/audit.Event
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
<h3 id="core.kjournal/v1alpha1.Event">Event
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
<code>apiVersion</code><br>
string</td>
<td>
<code>core.kjournal/v1alpha1</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
string
</td>
<td>
<code>Event</code>
</td>
</tr>
<tr>
<td>
<code>Event</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#event-v1-events">
Kubernetes events/v1.Event
</a>
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
<h3 id="core.kjournal/v1alpha1.Log">Log
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
<code>core.kjournal/v1alpha1</code>
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
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>payload</code><br>
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
<h3 id="core.kjournal/v1alpha1.ContainerLog">ContainerLog
</h3>
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
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
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
<code>payload</code><br>
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
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
