<h1>Source API reference</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#config.kjournal%2fv1alpha1">config.kjournal/v1alpha1</a>
</li>
</ul>
<h2 id="config.kjournal/v1alpha1">config.kjournal/v1alpha1</h2>
Resource Types:
<ul class="simple"></ul>
<h3 id="config.kjournal/v1alpha1.API">API
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.kjournal/v1alpha1.APIServerConfig">APIServerConfig</a>)
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
<code>resource</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>fieldMap</code><br>
<em>
map[string][]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>dropFields</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>filter</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>backend</code><br>
<em>
<a href="#config.kjournal/v1alpha1.ApiBackend">
ApiBackend
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>defaultTimeRange</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.kjournal/v1alpha1.APIServerConfig">APIServerConfig
</h3>
<p>APIServerConfig</p>
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
<code>backend</code><br>
<em>
<a href="#config.kjournal/v1alpha1.Backend">
Backend
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>apis</code><br>
<em>
<a href="#config.kjournal/v1alpha1.API">
[]API
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
<h3 id="config.kjournal/v1alpha1.ApiBackend">ApiBackend
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.kjournal/v1alpha1.API">API</a>)
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
<code>elasticsearch</code><br>
<em>
<a href="#config.kjournal/v1alpha1.ApiBackendElasticsearch">
ApiBackendElasticsearch
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
<h3 id="config.kjournal/v1alpha1.ApiBackendElasticsearch">ApiBackendElasticsearch
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.kjournal/v1alpha1.ApiBackend">ApiBackend</a>)
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
<code>index</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>refreshRate</code><br>
<em>
<a href="https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>timestampFields</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>bulkSize</code><br>
<em>
int64
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="config.kjournal/v1alpha1.Backend">Backend
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.kjournal/v1alpha1.APIServerConfig">APIServerConfig</a>)
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
<code>elasticsearch</code><br>
<em>
<a href="#config.kjournal/v1alpha1.BackendElasticsearch">
BackendElasticsearch
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
<h3 id="config.kjournal/v1alpha1.BackendElasticsearch">BackendElasticsearch
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.kjournal/v1alpha1.Backend">Backend</a>)
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
<code>url</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>tls</code><br>
<em>
<a href="#config.kjournal/v1alpha1.TLS">
TLS
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
<h3 id="config.kjournal/v1alpha1.TLS">TLS
</h3>
<p>
(<em>Appears on:</em>
<a href="#config.kjournal/v1alpha1.BackendElasticsearch">BackendElasticsearch</a>)
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
<code>allowInsecure</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>caCert</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>serverName</code><br>
<em>
string
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
