// Restores the placeholder Angular's build just deleted.
//
// The bundle is compiled into the Go binary with go:embed, which will
// not compile against a directory that does not exist - and Angular
// empties its output directory before writing, placeholder included. A
// fresh clone then fails with "pattern all:dist: no matching files
// found" on anything that touches the Go code before the frontend is
// built: go vet, go test, make build-api, every CI job.
//
// It has gone missing twice this way. This runs after every build, so
// it cannot go missing a third time.
import { writeFileSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const dist = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'internal', 'webui', 'dist');
writeFileSync(
  join(dist, '.gitkeep'),
  '# Keeps this directory in git so go:embed always has something to\n' +
    '# point at. The built frontend lands here and is ignored.\n',
);
