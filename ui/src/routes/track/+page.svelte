<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { trackCargo } from '$lib/api';
	import type { TrackingInfo } from '$lib/types';

	let trackingId = $state('');
	let result = $state<TrackingInfo | null>(null);
	let loading = $state(false);
	let error = $state('');

	onMount(() => {
		const qid = $page.url.searchParams.get('id');
		if (qid) {
			trackingId = qid;
			handleTrack();
		}
	});

	async function handleTrack(e?: SubmitEvent) {
		e?.preventDefault();
		if (!trackingId.trim()) return;

		loading = true;
		error = '';
		result = null;

		try {
			result = await trackCargo(trackingId.trim());
		} catch (err) {
			error = (err as Error).message;
		} finally {
			loading = false;
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

<h1>Track Cargo</h1>

<div class="card">
	<form onsubmit={handleTrack}>
		<div class="row">
			<div class="form-group" style="flex: 3;">
				<label for="tid">Tracking ID</label>
				<input
					type="text"
					id="tid"
					bind:value={trackingId}
					placeholder="Enter tracking ID"
					required
				/>
			</div>
			<div class="form-group" style="flex: 0; align-self: flex-end;">
				<button type="submit" class="btn btn-primary" disabled={loading}>
					{loading ? 'Tracking...' : 'Track'}
				</button>
			</div>
		</div>
	</form>
</div>

{#if error}
	<div class="error-box">{error}</div>
{/if}

{#if result}
	<div class="card">
		<h2>Status</h2>
		<dl class="detail-grid">
			<dt>Status</dt>
			<dd>
				{result.statusText}
				{#if result.isMisdirected}
					<span class="badge badge-danger" style="margin-left: 0.5rem;">Misdirected</span>
				{/if}
			</dd>
			<dt>Destination</dt>
			<dd>{result.destination}</dd>
			{#if result.eta}
				<dt>ETA</dt>
				<dd>{formatDate(result.eta)}</dd>
			{/if}
			{#if result.nextExpectedActivity}
				<dt>Next</dt>
				<dd>{result.nextExpectedActivity}</dd>
			{/if}
		</dl>
	</div>

	{#if result.events.length > 0}
		<div class="card" style="padding: 0; overflow: hidden;">
			<div style="padding: 1.25rem 1.25rem 0.5rem;">
				<h2>Handling History</h2>
			</div>
			<table>
				<thead>
					<tr>
						<th>Event</th>
						<th>Location</th>
						<th>Voyage</th>
						<th>Time</th>
						<th>Expected</th>
					</tr>
				</thead>
				<tbody>
					{#each result.events as evt}
						<tr>
							<td>{evt.description}</td>
							<td>{evt.location}</td>
							<td class="mono">{evt.voyageNumber || '-'}</td>
							<td>{formatDate(evt.completionTime)}</td>
							<td>
								{#if evt.isExpected}
									<span class="badge badge-success">Yes</span>
								{:else}
									<span class="badge badge-danger">No</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
{/if}
