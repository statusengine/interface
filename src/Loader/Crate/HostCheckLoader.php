<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Loader\Crate;

use Crate\PDO\PDO;
use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\HostCheckLoaderInterface;
use Statusengine\ValueObjects\HostCheckQueryOptions;

class HostCheckLoader implements HostCheckLoaderInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * HostCheckLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    /**
     * @param HostCheckQueryOptions $HostCheckQueryOptions
     * @return array
     */
    public function getHostchecks(HostCheckQueryOptions $HostCheckQueryOptions) {
        $fields = [
            'hostname',
            'is_hardstate',
            'output',
            'start_time',
            'state',
            'current_check_attempt',
            'max_check_attempts'
        ];
        $baseQuery = sprintf('SELECT %s FROM statusengine_hostchecks WHERE hostname=?', implode(',', $fields));

        if ($HostCheckQueryOptions->sizeOfStateFilter() > 0 && $HostCheckQueryOptions->sizeOfStateFilter() < 3) {
            $baseQuery = sprintf('%s AND state IN(%s)', $baseQuery, implode(',', $HostCheckQueryOptions->getStateFilter()));
        }

        if ($HostCheckQueryOptions->getOutputLike() != '') {
            $baseQuery = sprintf(' %s AND output ~* ? ', $baseQuery);
        }

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $HostCheckQueryOptions->getOrder(),
            $HostCheckQueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        $query->bindValue($i++, $HostCheckQueryOptions->getHostname());

        if ($HostCheckQueryOptions->getOutputLike() != '') {
            $like = sprintf('.*%s.*', $HostCheckQueryOptions->getOutputLike());
            $query->bindValue($i++, $like);
        }

        $query->bindValue($i++, $HostCheckQueryOptions->getLimit(), PDO::PARAM_INT);
        $query->bindValue($i++, $HostCheckQueryOptions->getOffset(), PDO::PARAM_INT);

        return $this->Backend->fetchAll($query);
    }

}
