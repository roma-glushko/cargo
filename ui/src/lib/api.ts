import type { Cargo, Itinerary, Location, TrackingInfo } from './types';

const BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		headers: { 'Content-Type': 'application/json' },
		...options
	});

	if (!res.ok) {
		const body = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(body.error || res.statusText);
	}

	const text = await res.text();
	if (!text) return undefined as T;
	return JSON.parse(text);
}

export function listCargos(): Promise<Cargo[]> {
	return request('/cargos');
}

export function getCargo(id: string): Promise<Cargo> {
	return request(`/cargos/${id}`);
}

export function bookCargo(origin: string, destination: string, arrivalDeadline: string): Promise<{ trackingId: string }> {
	return request('/cargos', {
		method: 'POST',
		body: JSON.stringify({ origin, destination, arrivalDeadline })
	});
}

export function requestRoutes(id: string): Promise<Itinerary[]> {
	return request(`/cargos/${id}/routes`);
}

export function assignRoute(id: string, legs: Itinerary['legs']): Promise<void> {
	return request(`/cargos/${id}/route`, {
		method: 'PUT',
		body: JSON.stringify({ legs })
	});
}

export function changeDestination(id: string, destination: string): Promise<void> {
	return request(`/cargos/${id}/destination`, {
		method: 'PUT',
		body: JSON.stringify({ destination })
	});
}

export function listLocations(): Promise<Location[]> {
	return request('/locations');
}

export function trackCargo(id: string): Promise<TrackingInfo> {
	return request(`/track/${id}`);
}

export function registerHandlingEvent(
	trackingIds: string[],
	type: string,
	unLocode: string,
	completionTime: string,
	voyageNumber?: string
): Promise<void> {
	return request('/handling', {
		method: 'POST',
		body: JSON.stringify({ trackingIds, type, unLocode, completionTime, voyageNumber: voyageNumber || '' })
	});
}
