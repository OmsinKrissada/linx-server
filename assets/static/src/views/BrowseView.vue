<template>
  <Card class="container max-w-4xl mx-auto">
    <CardHeader>
      <CardTitle>Browse</CardTitle>
      <CardDescription>Files their uploaders chose to list publicly.</CardDescription>
    </CardHeader>

    <CardContent>
      <div v-if="loading" class="flex flex-col gap-3" aria-busy="true">
        <Skeleton v-for="i in 3" :key="i" class="h-10 w-full" />
      </div>

      <Alert v-else-if="error" variant="destructive">
        <ErrorIcon />
        <AlertTitle>Couldn't load the file list.</AlertTitle>
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <p v-else-if="!files.length" class="py-8 text-center text-muted-foreground">
        No public files yet. Turn on <strong>Show in public list</strong> when uploading to add one.
      </p>

      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead class="text-right">Size</TableHead>
            <TableHead class="hidden sm:table-cell pl-6">Uploaded</TableHead>
            <TableHead class="hidden sm:table-cell">Expires</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="file in files" :key="file.filename">
            <TableCell class="whitespace-normal">
              <div class="flex items-center gap-2">
                <RouterLink :to="`/${file.filename}`" class="link break-all">
                  {{ file.original_name || file.filename }}
                </RouterLink>
                <span
                  v-if="file.original_name && file.original_name !== file.filename"
                  class="text-xs text-muted-foreground break-all"
                  >/{{ file.filename }}</span
                >
                <Tooltip v-if="file.protected">
                  <TooltipTrigger as-child>
                    <LockIcon
                      class="shrink-0 text-muted-foreground"
                      aria-label="Password protected"
                    />
                  </TooltipTrigger>
                  <TooltipContent>Password protected</TooltipContent>
                </Tooltip>
                <Tooltip v-if="file.burn_after_read">
                  <TooltipTrigger as-child>
                    <BurnIcon class="shrink-0 text-destructive" aria-label="Burn after reading" />
                  </TooltipTrigger>
                  <TooltipContent>Deleted after it is viewed once</TooltipContent>
                </Tooltip>
              </div>
            </TableCell>
            <TableCell class="text-right tabular-nums">{{
              formatBytes(Number(file.size))
            }}</TableCell>
            <TableCell class="hidden sm:table-cell pl-6">
              <UseTimeAgo v-slot="{ timeAgo }" :time="toDate(file.uploaded)">
                <time
                  :datetime="toDate(file.uploaded).toISOString()"
                  :title="toDate(file.uploaded).toLocaleString()"
                  >{{ timeAgo }}</time
                >
              </UseTimeAgo>
            </TableCell>
            <TableCell class="hidden sm:table-cell">
              <UseTimeAgo
                v-if="Number(file.expiry)"
                v-slot="{ timeAgo }"
                :time="toDate(file.expiry)"
              >
                <time
                  :datetime="toDate(file.expiry).toISOString()"
                  :title="toDate(file.expiry).toLocaleString()"
                  >{{ timeAgo }}</time
                >
              </UseTimeAgo>
              <span v-else class="text-muted-foreground">Never</span>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { UseTimeAgo } from "@vueuse/components";
import axios from "axios";
import { onMounted, ref } from "vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ApiPath } from "@/config/api.ts";
import { formatBytes } from "@/util/bytes.ts";
import ErrorIcon from "~icons/material-symbols/error-rounded";
import BurnIcon from "~icons/material-symbols/local-fire-department-rounded";
import LockIcon from "~icons/material-symbols/lock-rounded";

type PublicFile = {
  filename: string;
  original_name?: string;
  size: string;
  mimetype: string;
  expiry: string;
  uploaded: string;
  protected?: boolean;
  burn_after_read?: boolean;
};

const files = ref<PublicFile[]>([]);
const loading = ref(true);
const error = ref("");

const toDate = (unix: string) => new Date(Number(unix) * 1000);

onMounted(async () => {
  try {
    const res = await axios.get<PublicFile[]>(ApiPath("/api/browse"));
    files.value = res.data;
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
});
</script>
