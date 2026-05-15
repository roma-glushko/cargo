<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { listLocations, bookCargo } from '$lib/api';
	import type { Location } from '$lib/types';

	let locations = $state<Location[]>([]);
	let origin = $state('');
	let destination = $state('');
	let deadline = $state('');
	let error = $state('');
	let submitting = $state(false);

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
		submitting = true;

		try {
			const deadlineISO = new Date(deadline).toISOString();
			const result = await bookCargo(origin, destination, deadlineISO);
			goto(`/cargos/${result.trackingId}`);
		} catch (err) {
			error = (err as Error).message;
		} finally {
			submitting = false;
		}
	}
</script>

<h1>Book New Cargo</h1>

{#if error}
	<div class="error-box">{error}</div>
{/if}

<div class="card">
	<form onsubmit={handleSubmit}>
		<div class="row">
			<div class="form-group">
				<label for="origin">Origin</label>
				<select id="origin" bind:value={origin} required>
					<option value="" disabled>Select origin</option>
					{#each locations as loc}
						<option value={loc.unLocode}>{loc.name} ({loc.unLocode})</option>
					{/each}
				</select>
			</div>
			<div class="form-group">
				<label for="destination">Destination</label>
				<select id="destination" bind:value={destination} required>
					<option value="" disabled>Select destination</option>
					{#each locations as loc}
						<option value={loc.unLocode}>{loc.name} ({loc.unLocode})</option>
					{/each}
				</select>
			</div>
		</div>
		<div class="form-group">
			<label for="deadline">Arrival Deadline</label>
			<input type="datetime-local" id="deadline" bind:value={deadline} required />
		</div>
		<div class="actions">
			<button type="submit" class="btn btn-primary" disabled={submitting}>
				{submitting ? 'Booking...' : 'Book Cargo'}
			</button>
		</div>
	</form>
</div>
