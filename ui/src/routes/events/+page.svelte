<script lang="ts">
	import { onMount } from 'svelte';
	import { listLocations, registerHandlingEvent } from '$lib/api';
	import type { Location } from '$lib/types';

	let locations = $state<Location[]>([]);
	let trackingIds = $state('');
	let eventType = $state('RECEIVE');
	let location = $state('');
	let voyageNumber = $state('');
	let completionTime = $state('');
	let error = $state('');
	let success = $state('');
	let submitting = $state(false);

	const eventTypes = ['RECEIVE', 'LOAD', 'UNLOAD', 'CLAIM', 'CUSTOMS'];
	const requiresVoyage = $derived(eventType === 'LOAD' || eventType === 'UNLOAD');

	onMount(async () => {
		try {
			locations = await listLocations();
		} catch (e) {
			error = (e as Error).message;
		}
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		success = '';
		submitting = true;

		const ids = trackingIds
			.split(/[,\n]+/)
			.map((s) => s.trim())
			.filter(Boolean);

		if (ids.length === 0) {
			error = 'Enter at least one tracking ID';
			submitting = false;
			return;
		}

		try {
			const timeISO = new Date(completionTime).toISOString();
			await registerHandlingEvent(ids, eventType, location, timeISO, requiresVoyage ? voyageNumber : undefined);
			success = `Event registered for ${ids.length} cargo(s)`;
			trackingIds = '';
			voyageNumber = '';
		} catch (err) {
			error = (err as Error).message;
		} finally {
			submitting = false;
		}
	}
</script>

<h1>Register Handling Event</h1>

{#if error}
	<div class="error-box">{error}</div>
{/if}
{#if success}
	<div class="success-box">{success}</div>
{/if}

<div class="card">
	<form onsubmit={handleSubmit}>
		<div class="form-group">
			<label for="tracking-ids">Tracking IDs (one per line or comma-separated)</label>
			<input
				type="text"
				id="tracking-ids"
				bind:value={trackingIds}
				placeholder="ABC123, DEF456"
				required
			/>
		</div>

		<div class="row">
			<div class="form-group">
				<label for="event-type">Event Type</label>
				<select id="event-type" bind:value={eventType} required>
					{#each eventTypes as t}
						<option value={t}>{t}</option>
					{/each}
				</select>
			</div>
			<div class="form-group">
				<label for="loc">Location</label>
				<select id="loc" bind:value={location} required>
					<option value="" disabled>Select location</option>
					{#each locations as loc}
						<option value={loc.unLocode}>{loc.name} ({loc.unLocode})</option>
					{/each}
				</select>
			</div>
		</div>

		{#if requiresVoyage}
			<div class="form-group">
				<label for="voyage">Voyage Number</label>
				<input
					type="text"
					id="voyage"
					bind:value={voyageNumber}
					placeholder="e.g. V100"
					required
				/>
			</div>
		{/if}

		<div class="form-group">
			<label for="comp-time">Completion Time</label>
			<input type="datetime-local" id="comp-time" bind:value={completionTime} required />
		</div>

		<div class="actions">
			<button type="submit" class="btn btn-primary" disabled={submitting}>
				{submitting ? 'Registering...' : 'Register Event'}
			</button>
		</div>
	</form>
</div>
