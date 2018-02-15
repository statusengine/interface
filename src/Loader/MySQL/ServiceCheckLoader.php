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

namespace Statusengine\Loader\Mysql;

use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\ServiceCheckLoaderInterface;
use Statusengine\ValueObjects\ServiceCheckQueryOptions;

class ServiceCheckLoader implements ServiceCheckLoaderInterface {

    /**
     * @var \Statusengine\Backend\Mysql\MySQL
     */
    private $Backend;

    /**
     * ServiceCheckLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    /**
     * @param ServiceCheckQueryOptions $ServiceCheckQueryOptions
     * @return array
     */
    public function getServicechecks(ServiceCheckQueryOptions $ServiceCheckQueryOptions) {
        $fields = [
            'booleans' => [
                'is_hardstate'
            ],
            'strings' => [
                'hostname',
                'service_description',
                'output',
                'state',
                'current_check_attempt',
                'max_check_attempts',
                'start_time'
            ]
        ];

        $sql = [];
        foreach ($fields['booleans'] as $field) {
            $sql[] = $field;
        }
        foreach ($fields['strings'] as $field) {
            $sql[] = $field;
        }


        $baseQuery = sprintf('SELECT %s FROM statusengine_servicechecks WHERE hostname=? AND service_description=?', implode(',', $sql));

        if ($ServiceCheckQueryOptions->sizeOfStateFilter() > 0 && $ServiceCheckQueryOptions->sizeOfStateFilter() < 4) {
            $baseQuery = sprintf('%s AND state IN(%s)', $baseQuery, implode(',', $ServiceCheckQueryOptions->getStateFilter()));
        }

        if ($ServiceCheckQueryOptions->getOutputLike() != '') {
            $baseQuery = sprintf(' %s AND output LIKE ? ', $baseQuery);
        }

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $ServiceCheckQueryOptions->getOrder(),
            $ServiceCheckQueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        $query->bindValue($i++, $ServiceCheckQueryOptions->getHostname());
        $query->bindValue($i++, $ServiceCheckQueryOptions->getServiceDescription());

        if ($ServiceCheckQueryOptions->getOutputLike() != '') {
            $like = sprintf('%%%s%%', $ServiceCheckQueryOptions->getOutputLike());
            $query->bindValue($i++, $like);
        }

        $query->bindValue($i++, $ServiceCheckQueryOptions->getLimit(), \PDO::PARAM_INT);
        $query->bindValue($i++, $ServiceCheckQueryOptions->getOffset(), \PDO::PARAM_INT);

        $results = $this->Backend->fetchAll($query);

        foreach($results as $key => $result){
            foreach ($fields['booleans'] as $field) {
                $results[$key][$field] = (bool)$results[$key][$field];
            }
        }

        return $results;

    }

}
