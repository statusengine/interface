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
use Statusengine\Loader\ServiceNotificationLoaderInterface;
use Statusengine\ValueObjects\ServiceNotificationQueryOptions;

class ServiceNotificationLoader implements ServiceNotificationLoaderInterface {


    /**
     * @var \Statusengine\Backend\Mysql\MySQL
     */
    private $Backend;

    /**
     * ServiceNotificationLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    /**
     * @param ServiceNotificationQueryOptions $ServiceNotificationQueryOptions
     * @return array
     */
    public function getServiceNotifications(ServiceNotificationQueryOptions $ServiceNotificationQueryOptions) {
        $fields = [
            'strings' => [
                'hostname',
                'output',
                'state',
                'command_name',
                'contact_name',
                'reason_type',
                'start_time'
            ]
        ];

        $sql = [];
        foreach ($fields['strings'] as $field) {
            $sql[] = $field;
        }


        $baseQuery = sprintf(
            'SELECT %s FROM statusengine_service_notifications WHERE hostname=? AND service_description=?',
            implode(',', $sql)
        );

        if ($ServiceNotificationQueryOptions->sizeOfStateFilter() > 0 && $ServiceNotificationQueryOptions->sizeOfStateFilter() < 4) {
            $baseQuery = sprintf('%s AND state IN(%s)', $baseQuery, implode(',', $ServiceNotificationQueryOptions->getStateFilter()));
        }

        if ($ServiceNotificationQueryOptions->getOutputLike() != '') {
            $baseQuery = sprintf(' %s AND output LIKE ? ', $baseQuery);
        }

        if ($ServiceNotificationQueryOptions->hasReasonTypeFilter()) {
            $baseQuery = sprintf(' %s reason_type=?', $baseQuery);
        }

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $ServiceNotificationQueryOptions->getOrder(),
            $ServiceNotificationQueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        $query->bindValue($i++, $ServiceNotificationQueryOptions->getHostname());
        $query->bindValue($i++, $ServiceNotificationQueryOptions->getServiceDescription());


        if ($ServiceNotificationQueryOptions->hasReasonTypeFilter()) {
            $query->bindValue($i++, $ServiceNotificationQueryOptions->getReasonType());
        }

        if ($ServiceNotificationQueryOptions->getOutputLike() != '') {
            $like = sprintf('%%%s%%', $ServiceNotificationQueryOptions->getOutputLike());
            $query->bindValue($i++, $like);
        }

        $query->bindValue($i++, $ServiceNotificationQueryOptions->getLimit(), \PDO::PARAM_INT);
        $query->bindValue($i++, $ServiceNotificationQueryOptions->getOffset(), \PDO::PARAM_INT);

        return $this->Backend->fetchAll($query);
    }

}
