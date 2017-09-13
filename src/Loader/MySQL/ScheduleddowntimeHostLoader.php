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

namespace Statusengine\Loader\Mysql;

use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\ScheduleddowntimeHostLoaderInterface;
use Statusengine\ValueObjects\ScheduleddowntimeQueryOptions;

class ScheduleddowntimeHostLoader implements ScheduleddowntimeHostLoaderInterface {

    /**
     * @var \Statusengine\Backend\Mysql\MySQL
     */
    private $Backend;

    /**
     * ScheduleddowntimeHostLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    public function getScheduledHostdowntimes(ScheduleddowntimeQueryOptions $QueryOptions) {
        $fields = [
            'booleans' => [
                'was_started'
            ],
            'strings' => [
                'hostname',
                'author_name',
                'comment_data',
                'internal_downtime_id',
                'scheduled_start_time',
                'scheduled_end_time',
                'duration',
                'node_name'
            ]
        ];

        $sql = [];
        foreach ($fields['booleans'] as $field) {
            $sql[] = $field;
        }
        foreach ($fields['strings'] as $field) {
            $sql[] = $field;
        }


        $baseQuery = sprintf('SELECT %s FROM statusengine_host_scheduleddowntimes WHERE 1=1 ', implode(',', $sql));

        if ($QueryOptions->getHostnameLike() != '') {
            $baseQuery = sprintf('%s AND hostname LIKE ?', $baseQuery);
        }

        $baseQuery .= $this->getClusterNameQuery($QueryOptions);

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $QueryOptions->getOrder(),
            $QueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        if ($QueryOptions->getHostnameLike() != '') {
            $like = sprintf('%%%s%%', $QueryOptions->getHostnameLike());
            $query->bindValue($i++, $like);
        }

        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $query->bindValue($i++, $clusterName);
        }

        $query->bindValue($i++, $QueryOptions->getLimit(), \PDO::PARAM_INT);
        $query->bindValue($i++, $QueryOptions->getOffset(), \PDO::PARAM_INT);

        $results = $this->Backend->fetchAll($query);

        foreach ($results as $key => $result) {
            foreach ($fields['booleans'] as $field) {
                $results[$key][$field] = (bool)$results[$key][$field];
            }
        }

        return $results;

    }

    /**
     * @param ScheduleddowntimeQueryOptions $QueryOptions
     * @return string
     */
    private function getClusterNameQuery(ScheduleddowntimeQueryOptions $QueryOptions) {
        $placeholders = [];
        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $placeholders[] = '?';
        }
        if (!empty($placeholders)) {
            return sprintf(' AND node_name IN(%s)', implode(',', $placeholders));
        }
        return '';
    }

}
