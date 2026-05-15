<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { listCargos } from '$lib/api';
	import type { Cargo } from '$lib/types';

	let cargos = $state<Cargo[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			cargos = await listCargos();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			loading = false;
		}
	});

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}
</script>

<h1>Cargos</h1>

{#if error}
	<div class="error-box">{error}</div>
{/if}

{#if loading}
	<p class="spinner">Loading cargos...</p>
{:else}
	<div class="card" style="padding: 0; overflow: hidden;">
		<table>
			<thead>
				<tr>
					<th>Tracking ID</th>
					<th>Origin</th>
					<th>Destination</th>
					<th>Deadline</th>
					<th>Status</th>
				</tr>
			</thead>
			<tbody>
				{#each cargos as cargo}
					<tr onclick={() => goto(`/cargos/${cargo.trackingId}`)} style="cursor: pointer;">
						<td><span class="mono">{cargo.trackingId}</span></td>
						<td>{cargo.origin}</td>
						<td>{cargo.destination}</td>
						<td>{formatDate(cargo.arrivalDeadline)}</td>
						<td>
							{#if cargo.misrouted}
								<span class="badge badge-danger">Misrouted</span>
							{:else if cargo.routed}
								<span class="badge badge-success">Routed</span>
							{:else}
								<span class="badge badge-warning">Not Routed</span>
							{/if}
						</td>
					</tr>
				{:else}
					<tr>
						<td colspan="5" style="text-align: center; color: var(--color-text-muted);">
							No cargos found. <a href="/book">Book one</a>.
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
