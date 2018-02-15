<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler

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
use Statusengine\Loader\ServiceStateHistoryLoaderInterface;
use Statusengine\ValueObjects\ServiceStateHistoryQueryOptions;

class ServiceStateHistoryLoader implements ServiceStateHistoryLoaderInterface {


    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * ServiceStateHistoryLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }


    public function getServiceStateHistory(ServiceStateHistoryQueryOptions $ServiceStateHistoryQueryOptions) {
        $fields = [
            'hostname',
            'service_description',
            'is_hardstate',
            'output',
            'state_time',
            'state',
            'current_check_attempt',
            'max_check_attempts'
        ];
        $baseQuery = sprintf('SELECT %s FROM  statusengine_service_statehistory  WHERE hostname=? AND service_description=?', implode(',', $fields));

        if ($ServiceStateHistoryQueryOptions->sizeOfStateFilter() > 0 && $ServiceStateHistoryQueryOptions->sizeOfStateFilter() < 4) {
            $baseQuery = sprintf('%s AND state IN(%s)', $baseQuery, implode(',', $ServiceStateHistoryQueryOptions->getStateFilter()));
        }

        if ($ServiceStateHistoryQueryOptions->getOutputLike() != '') {
            $baseQuery = sprintf(' %s AND output ~* ? ', $baseQuery);
        }

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $ServiceStateHistoryQueryOptions->getOrder(),
            $ServiceStateHistoryQueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        $query->bindValue($i++, $ServiceStateHistoryQueryOptions->getHostname());
        $query->bindValue($i++, $ServiceStateHistoryQueryOptions->getServiceDescription());

        if ($ServiceStateHistoryQueryOptions->getOutputLike() != '') {
            $like = sprintf('.*%s.*', $ServiceStateHistoryQueryOptions->getOutputLike());
            $query->bindValue($i++, $like);
        }

        $query->bindValue($i++, $ServiceStateHistoryQueryOptions->getLimit(), PDO::PARAM_INT);
        $query->bindValue($i++, $ServiceStateHistoryQueryOptions->getOffset(), PDO::PARAM_INT);

        return $this->Backend->fetchAll($query);
    }

}
