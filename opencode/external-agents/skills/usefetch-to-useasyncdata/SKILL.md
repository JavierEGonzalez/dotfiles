---
name: usefetch-to-useasyncdata
description: >-
  Refactor Nuxt useFetch calls to useAsyncData with $fetch in this Prism3 monorepo.
  Use when asked to migrate useFetch, convert data fetching to useAsyncData, or make
  fetched data directly accessible in a template without .data.value nesting.
---

# useFetch to useAsyncData Refactor

## When to Apply

Use this skill when:
- User asks to "use useAsyncData", "convert useFetch", "refactor data fetching"
- User wants fetched data directly accessible in the template
- Multiple sequential/conditional `useFetch` calls need consolidation
- Data needs to be composed from multiple API calls into a single reactive object
- Multiple **independent** fetches can run in parallel and partial failure should be tolerated

# Required Every Refactor
ALWAYS have the useAsyncData return data. 
NEVER have side effects that happen inside the useAsyncData.
NEVER call useAsyncData inside a control flow statment. The return values should be accesible in the root.

## Pattern 1: Single useFetch (simple replacement)

**Before:**
```ts
const { data, error } = await useFetch<MyType>(ENDPOINT, {
  key: 'my-key',
  query: { path: 'some/path' },
  ...useFetchWithCache(),
})
if (!data.value?.documents || error.value) throw new Error('...')
const result = data.value.documents[0]
```

In this case respond to the user the following:
[useFetch correctly used]

## Pattern 2: Multiple/conditional fetches composed into one

**Before:**
```ts
const { data: mappingData } = await useFetch(ENDPOINT, { key: 'mapping', query: {...} })
if (mappingData.value) {
  const { data: detailData } = await useFetch(ENDPOINT, { key: dynamicKey, query: {...} })
  // process detailData
}
```

**After:**
```ts
const { data, error } = await useAsyncData('mapping', async () => {
  const mapping = await $fetch<MappingType>(ENDPOINT, { query: {...} })
  if (!mapping.documents?.[0]) return { mapping: null, detail: null }

  const detail = await $fetch<DetailType>(ENDPOINT, {
    query: { path: mapping.documents[0].fields.someDoc.path }
  })
  return { mapping, detail }
}, { server: true })

// data.value is now { mapping, detail } — directly usable in template
```

## Pattern 3: Complex control flow — extract to a composable-style async function

When the fetch logic has many conditionals, error handling branches, or parser steps, extract it into an async function called from `<script setup>`. This keeps the setup block clean and the logic testable.

**Before (sprawling try/catch with multiple useFetch calls in setup):**
```ts
const pillsData = ref<NavLinkItem[]>([])
const kwmsData = ref<BrKwmData[]>([])
const hasMapping = ref(false)

try {
  const { data, error } = await useFetch(ENDPOINT, { key: 'mapping', ... })
  if (!data.value || error.value) throw new Error('...')

  const match = data.value.documents[0]?.fields.find(...)
  if (match) {
    hasMapping.value = true
    const results = await useFetch(ENDPOINT, { key: match.name, ... })
    // ... 40 lines of parsing and validation ...
    pillsData.value = parsedPills
    kwmsData.value = parsedKwms
  }
} catch (e) {
  hasMapping.value = false
  console.error(e)
}
```

**After (extracted async function with useAsyncData + $fetch inside):**
```ts
interface SearchTermData {
  hasMapping: boolean
  pills: NavLinkItem[]
  kwms: BrKwmData[]
  recipePayload: RecipeCarouselPayload | null
  regularTiles: BrDocument<BrRegularTileFieldset>[]
}

const emptyResult: SearchTermData = {
  hasMapping: false, pills: [], kwms: [], recipePayload: null, regularTiles: [],
}

const useSearchTermMapping = async () => {
  const { data, error } = await useAsyncData<SearchTermData>('search-term-mapping', async () => {
    let mapping: BrDocumentResponse<BrSearchTermMappingResponse> = await $fetch<BrDocumentResponse<BrSearchTermMappingResponse>>(
      ENDPOINT, { query: { path: `${brBrandChannel}/search/search-term-mapping` } },
    )

    if (!mapping.documents?.[0]) {
      return emptyResult
    }

    const match = mapping.documents[0].fields.searchTermDocuments.find(...)
    if (!match) {
      const tilePool = await $fetch<BrDocumentResponse<BrRegularTileFieldset>>(
        ENDPOINT, { query: { folder: `${brBrandChannel}/search/regular-tiles` } },
      )
      return { ...emptyResult, regularTiles: tilePool.documents ?? [] }
    }

    // Match found — fetch the mapped document
    const resultDoc = await $fetch<BrSearchTermResultDocumentResponse>(
      ENDPOINT, { query: { path: match.searchTermDocument.path } },
    )

    if (!resultDoc.documents?.[0]) {
      throw new Error('Search term result document contains no documents.')
    }

    const fields = resultDoc.documents[0].fields
    return {
      hasMapping: true,
      pills: parsePillsData(fields?.pills ?? []),
      kwms: parseKeyWebMessagesData(fields?.kwmGroup?.keyWebMessageCarousel ?? []),
      recipePayload: parseRecipeCarouselPayload(fields?.recipeCarousel),
      regularTiles: [],
    }
  }, { server: true })

  return { data, error }
}

const { data: searchTermData, error: searchTermError } = await useSearchTermMapping()

if (searchTermError.value) {
  console.error('Search term mapping fetch failed:', searchTermError.value)
}

// Derive template-bound values via computed
const hasSearchTermMapping = computed(() => searchTermData.value?.hasMapping ?? false)
const pillsData = computed(() => searchTermData.value?.pills ?? [])
const kwmsData = computed(() => searchTermData.value?.kwms ?? [])
const recipeCarouselPayload = computed(() => searchTermData.value?.recipePayload ?? null)
const regularTilePool = computed(() => searchTermData.value?.regularTiles ?? [])
```

**Key benefits of this pattern:**
- All fetch + parse logic is in one function
- Template reads from computed properties derived from a single `useAsyncData` result
- Error handling is centralized
- Easy to test by mocking `$fetch` responses

## Pattern 4: Multiple independent fetches with Promise.allSettled

Use when several `useFetch` calls are **independent** (none depends on another's result) and **partial failure is acceptable** — the page should still render even if one request fails.

**Before:**
```ts
const { data: heroData } = await useFetch<HeroType>(ENDPOINT, { key: 'hero', query: {...} })
const { data: promoData } = await useFetch<PromoType>(ENDPOINT, { key: 'promo', query: {...} })
const { data: tileData } = await useFetch<TileType>(ENDPOINT, { key: 'tiles', query: {...} })
```

**After:**
```ts
interface PageData {
  hero: HeroType | null
  promo: PromoType | null
  tiles: TileType | null
}

const { data, error } = await useAsyncData<PageData>('page-data', async () => {
  const [heroResult, promoResult, tileResult] = await Promise.allSettled([
    $fetch<HeroType>(ENDPOINT, { query: { path: 'brand/hero' } }),
    $fetch<PromoType>(ENDPOINT, { query: { path: 'brand/promo' } }),
    $fetch<TileType>(ENDPOINT, { query: { folder: 'brand/tiles' } }),
  ])

  if (heroResult.status === 'rejected') console.error('Hero fetch failed:', heroResult.reason)
  if (promoResult.status === 'rejected') console.error('Promo fetch failed:', promoResult.reason)
  if (tileResult.status === 'rejected') console.error('Tiles fetch failed:', tileResult.reason)

  return {
    hero: heroResult.status === 'fulfilled' ? heroResult.value : null,
    promo: promoResult.status === 'fulfilled' ? promoResult.value : null,
    tiles: tileResult.status === 'fulfilled' ? tileResult.value : null,
  }
}, { server: true })

const heroData = computed(() => data.value?.hero ?? null)
const promoData = computed(() => data.value?.promo ?? null)
const tileData = computed(() => data.value?.tiles ?? null)
```

**Key benefits:**
- All independent requests fire in parallel — no waterfall
- A single failed fetch does not throw or suppress the others
- Each result's `.status` (`'fulfilled'` | `'rejected'`) lets you log/handle failures individually
- Template still renders with whatever data succeeded

**When to prefer `Promise.allSettled` over `Promise.all`:**
- Use `Promise.allSettled` when partial data is useful (non-critical sections can be empty/hidden)
- Use `Promise.all` when all data is required and one failure should abort the whole operation

### Spec File: mocking Promise.allSettled

```ts
const fetchMock = vi.hoisted(() => vi.fn())
vi.stubGlobal('$fetch', fetchMock)
vi.stubGlobal('useAsyncData', vi.fn(async (_key: string, handler: () => Promise<unknown>) => {
  try {
    const data = ref(await handler())
    return { data, error: ref(null) }
  } catch (e) {
    return { data: ref(null), error: ref(e) }
  }
}))

// Simulate one fetch failing:
fetchMock.mockImplementation((_url, options) => {
  if (options?.query?.path?.includes('hero')) return Promise.resolve(heroApiResponse)
  if (options?.query?.path?.includes('promo')) return Promise.reject(new Error('Promo unavailable'))
  if (options?.query?.folder?.includes('tiles')) return Promise.resolve(tileApiResponse)
})

// Assert partial result:
expect(heroData.value).toEqual(parsedHero)
expect(promoData.value).toBeNull()
expect(tileData.value).toEqual(parsedTiles)
```

## Caching

`useFetchWithCache()` provided `transform` + `getCachedData` to `useFetch`. With `useAsyncData`, pass `getCachedData` directly in the options:

```ts
const { data } = await useAsyncData('my-key',
  () => $fetch<MyType>(ENDPOINT, { query }),
  {
    server: true,
    getCachedData(key, nuxtApp) {
      const cached = nuxtApp.payload.data[key] ?? nuxtApp.static.data[key]
      if (!cached) return null
      const isExpired = (cached.fetchedAt + 1_800_000) < Date.now()
      return isExpired ? null : cached
    },
  },
)
```

If caching is not needed, omit `getCachedData` entirely.

## Spec File Updates

### Mocking strategy

Replace `useFetch` stubs with `useAsyncData` + `$fetch` stubs:

**Before:**
```ts
const useFetchMock = vi.hoisted(() => vi.fn())
vi.stubGlobal('useFetch', useFetchMock)
vi.stubGlobal('useFetchWithCache', vi.fn(() => ({})))

useFetchMock.mockImplementation((endpoint, options) => {
  if (options?.key === 'my-key') return myMock(endpoint, options)
  return otherMock(endpoint, options)
})
```

**After:**
```ts
const fetchMock = vi.hoisted(() => vi.fn())
vi.stubGlobal('$fetch', fetchMock)
vi.stubGlobal('useAsyncData', vi.fn(async (_key: string, handler: () => Promise<unknown>) => {
  try {
    const data = ref(await handler())
    return { data, error: ref(null) }
  } catch (e) {
    return { data: ref(null), error: ref(e) }
  }
}))

// Control behaviour by mocking $fetch per URL/query:
fetchMock.mockImplementation((_url, options) => {
  if (options?.query?.path?.includes('search-term-mapping')) {
    return Promise.resolve(mappingApiResponse)
  }
  if (options?.query?.folder?.includes('regular-tiles')) {
    return Promise.resolve(tilePoolApiResponse)
  }
  return Promise.resolve(detailApiResponse)
})
```

### Key differences in test assertions

- No more asserting on `useFetchMock` calls with `key` options
- Assert on `$fetch` calls with URL and query params
- Error scenarios: make `fetchMock` reject or return empty data

## Refactor Checklist

1. Read the file and identify all `useFetch` calls and their keys
2. Identify which fetches are sequential/conditional vs. independent
3. Choose the appropriate pattern:
   - **Pattern 1** inform the user it is currently used correctly.
   - **Pattern 2** for 2-3 related fetches that can be grouped
   - **Pattern 3** for complex control flow with many branches/parsers
   - **Pattern 4** for multiple independent fetches where partial failure is tolerable — use `Promise.allSettled`
4. Group related fetches into one `useAsyncData` callback
5. Return a structured object from the callback
6. Replace imperative `ref` assignments with `computed` derived from `data.value`
7. Remove `useFetchWithCache()` spreads (add inline `getCachedData` only if caching is required)
8. Update spec: replace `useFetch` stub with `useAsyncData` + `$fetch` stubs
9. Run tests and lint

## Verification

```bash
# Run tests for the changed file
pnpm --filter prism test:unit <path/to/file.spec.ts>

# Run lint
pnpm --filter prism lint <path/to/file.vue> <path/to/file.spec.ts>
```
