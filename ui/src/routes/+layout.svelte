<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	let { children } = $props();

	const links = [
		{ href: '/', label: 'Cargos' },
		{ href: '/book', label: 'Book Cargo' },
		{ href: '/track', label: 'Track' },
		{ href: '/events', label: 'Register Event' }
	];
</script>

<div class="shell">
	<nav>
		<div class="brand">Cargo Admin</div>
		{#each links as link}
			<a href={link.href} class:active={$page.url.pathname === link.href}>{link.label}</a>
		{/each}
	</nav>
	<main>
		{@render children()}
	</main>
</div>
