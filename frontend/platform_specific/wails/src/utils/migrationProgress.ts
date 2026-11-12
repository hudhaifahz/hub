import { EventsOn } from "wailsjs/runtime/runtime";

export type MigrationProgress = {
  step: string;
  detail: string;
};

export function onMigrationProgress(
  callback: (progress: MigrationProgress) => void
) {
  return EventsOn("grynvault:migration-progress", callback);
}
