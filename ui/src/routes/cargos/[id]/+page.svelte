<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import {
		getCargo,
		requestRoutes,
		assignRoute,
		changeDestination,
		listLocations
	} from '$lib/api';
	import type { Cargo, Itinerary, Location } from '$lib/types';

	let cargo = $state<Cargo | null>(null);
	let loading = $state(true);
	let error = $state('');
	let success = $state('');

	let routes = $state<Itinerary[]>([]);
	let loadingRoutes = $state(false);
	let showRoutes = $state(false);

	let showDestForm = $state(false);
	let locations = $state<Location[]>([]);
	let newDestination = $state('');

	const id = $derived($page.params.id);

	onMount(() => loadCargo());

	async function loadCargo() {
		loading = true;
		error = '';
		try {
			cargo = await getCargo(id);
		} catch (e) {
			error = (e as Error).message;
		} finally {
			loading = false;
		}
	}

	async function handleRequestRoutes() {
		loadingRoutes = true;
		error = '';
		try {
			routes = await requestRoutes(id);
			showRoutes = true;
		} catch (e) {
			error = (e as Error).message;
		} finally {
			loadingRoutes = false;
		}
	}

	async function handleAssignRoute(itin: Itinerary) {
		error = '';
		try {
			await assignRoute(id, itin.legs);
			success = 'Route assigned successfully';
			showRoutes = false;
			routes = [];
			await loadCargo();
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function handleOpenDestForm() {
		if (locations.length === 0) {
			locations = await listLocations();
		}
		showDestForm = true;
	}

	async function handleChangeDestination(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		try {
			await changeDestination(id, newDestination);
			success = 'Destination changed';
			showDestForm = false;
			newDestination = '';
			await loadCargo();
		} catch (err) {
			error = (err as Error).message;
		}
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

{#if loading}
	<p class="spinner">Loading...</p>
{:else if cargo}
	<h1>
		Cargo <span class="mono">{cargo.trackingId}</span>
	</h1>

	{#if error}
		<div class="error-box">{error}</div>
	{/if}
	{#if success}
		<div class="success-box">{success}</div>
	{/if}

	<div class="card">
		<h2>Details</h2>
		<dl class="detail-grid">
			<dt>Origin</dt>
			<dd>{cargo.origin}</dd>
			<dt>Destination</dt>
			<dd>{cargo.destination}</dd>
			<dt>Deadline</dt>
			<dd>{formatDate(cargo.arrivalDeadline)}</dd>
			<dt>Routing</dt>
			<dd>
				{#if cargo.misrouted}
					<span class="badge badge-danger">Misrouted</span>
				{:else if cargo.routed}
					<span class="badge badge-success">Routed</span>
				{:else}
					<span class="badge badge-warning">Not Routed</span>
				{/if}
			</dd>
		</dl>

		<div class="actions">
			{#if !cargo.routed}
				<button class="btn btn-primary" onclick={handleRequestRoutes} disabled={loadingRoutes}>
					{loadingRoutes ? 'Loading...' : 'Request Routes'}
				</button>
			{/if}
			<button class="btn btn-secondary" onclick={handleOpenDestForm}>
				Change Destination
			</button>
			<a href="/track?id={cargo.trackingId}" class="btn btn-secondary">
				Track
			</a>
		</div>
	</div>

	{#if showDestForm}
		<div class="card">
			<h2>Change Destination</h2>
			<form onsubmit={handleChangeDestination}>
				<div class="form-group">
					<label for="new-dest">New Destination</label>
					<select id="new-dest" bind:value={newDestination} required>
						<option value="" disabled>Select destination</option>
						{#each locations as loc}
							<option value={loc.unLocode}>{loc.name} ({loc.unLocode})</option>
						{/each}
					</select>
				</div>
				<div class="actions">
					<button type="submit" class="btn btn-primary">Save</button>
					<button type="button" class="btn btn-secondary" onclick={() => (showDestForm = false)}>
						Cancel
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if showRoutes && routes.length > 0}
		<div class="card">
			<h2>Candidate Routes ({routes.length})</h2>
			{#each routes as itin, idx}
				<div style="margin-bottom: 1rem; padding: 0.75rem; background: #f8fafc; border-radius: var(--radius);">
					<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">
						<strong>Route {idx + 1}</strong>
						<button class="btn btn-primary btn-sm" onclick={() => handleAssignRoute(itin)}>
							Assign
						</button>
					</div>
					<table>
						<thead>
							<tr>
								<th>Voyage</th>
								<th>From</th>
								<th>To</th>
								<th>Load</th>
								<th>Unload</th>
							</tr>
						</thead>
						<tbody>
							{#each itin.legs as leg}
								<tr>
									<td class="mono">{leg.voyageNumber}</td>
									<td>{leg.from}</td>
									<td>{leg.to}</td>
									<td>{formatDate(leg.loadTime)}</td>
									<td>{formatDate(leg.unloadTime)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/each}
		</div>
	{/if}

	{#if cargo.legs && cargo.legs.length > 0}
		<div class="card">
			<h2>Itinerary</h2>
			<table>
				<thead>
					<tr>
						<th>Voyage</th>
						<th>From</th>
						<th>To</th>
						<th>Load</th>
						<th>Unload</th>
					</tr>
				</thead>
				<tbody>
					{#each cargo.legs as leg}
						<tr>
							<td class="mono">{leg.voyageNumber}</td>
							<td>{leg.from}</td>
							<td>{leg.to}</td>
							<td>{formatDate(leg.loadTime)}</td>
							<td>{formatDate(leg.unloadTime)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
{:else}
	<div class="error-box">Cargo not found</div>
{/if}
