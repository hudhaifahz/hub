export type MigrationProgress = {
  step: string;
  detail: string;
};

export function onMigrationProgress(
  _callback: (progress: MigrationProgress) => void
) {
  return () => undefined;
}
